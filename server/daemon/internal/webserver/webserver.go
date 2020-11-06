package webserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/daemon/internal/monitor"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"io"
	core "k8s.io/api/core/v1"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var reverseProxy *httputil.ReverseProxy

var logger *log.Logger

func StartWebserver(parent context.Context) error {

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "webserver.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	logger.Println("Starting webserver")

	if err := generateKeys(); err != nil {
		return err
	}

	ui.Println("Starting webserver...")
	defer ui.Println("Starting webserver..DONE.")

	http.HandleFunc("/", handleRequest)

	go func() {
		logger.Println("Starting listener on port 80")
		if err := http.ListenAndServe("127.0.0.1:80", nil); err != nil {
			e := fmt.Errorf("error starting webserver listener on port 80: %s", err)
			logger.Println(e)
			//ui.Fatal(e)
		}
	}()

	go func() {
		logger.Println("Starting listener on port 443")
		if err := http.ListenAndServeTLS("127.0.0.1:443",
			environment.ServerHome+"/data/ssl-cert.pem",
			environment.ServerHome+"/data/ssl-key.pem",
			nil); err != nil {
			e := fmt.Errorf("error starting webserver listener on port 443: %s", err)
			logger.Println(e)
			//ui.Fatal(e)
		}
	}()

	go func() {
		select {
		case <-parent.Done():
			logger.Println("Stopping webserver...")
			logger.Println("Stopping webserver...DONE")
		}
	}()

	//go watchTraefikService()

	return nil
}

var traefikIp string

//func watchTraefikService() {
//	for !kubeclient.ConfigExists() {
//		logger.Printf("Webserver waiting for %s", kubeclient.KubeconfigFile)
//
//		time.Sleep(10 * time.Second)
//	}
//
//	var err error
//	kubeClient, err := kubeclient.KubeClient(environment.ServerHome)
//	if err != nil {
//		logger.Printf("ERROR: %s", err)
//	}
//
//	factory := informers.NewSharedInformerFactory(kubeClient, 0)
//	informer := factory.Core().V1().Services().Informer()
//	stopper := make(chan struct{})
//	defer close(stopper)
//
//	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
//		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
//			newService := newObj.(*core.Service)
//
//			checkTraefikService(newService)
//		},
//
//		AddFunc: func(obj interface{}) {
//			checkTraefikService(obj.(*core.Service))
//
//		},
//
//		DeleteFunc: func(obj interface{}) {
//			delService := obj.(*core.Service)
//
//			if delService.ObjectMeta.Namespace == "kube-system" && delService.ObjectMeta.Name == "traefik" {
//				traefikIp = ""
//				reverseProxy = nil
//				logger.Printf("Traefik service removed")
//			}
//		},
//	})
//	informer.Run(stopper)
//
//}

func checkTraefikService(newService *core.Service) {
	if newService.ObjectMeta.Namespace == "kube-system" && newService.ObjectMeta.Name == "traefik" {
		if traefikIp != newService.Spec.ClusterIP {
			traefikIp = newService.Spec.ClusterIP

			logger.Printf("Traefik IP is now %s. Configuring proxy...", traefikIp)

			internalUrl, err := url.Parse(fmt.Sprintf("http://%s", traefikIp))
			if err != nil {
				logger.Printf("ERROR: %s", err)
			}

			reverseProxy = httputil.NewSingleHostReverseProxy(internalUrl)
			reverseProxy.ErrorHandler = func(response http.ResponseWriter, request *http.Request, err error) {
				if err.Error() == "Gateway Error" {
					if err := showSiteDownPage(response); err != nil {
						logger.Printf("ERROR: %s", err)
					}
				}
			}

			reverseProxy.ModifyResponse = func(response *http.Response) error {
				if response.StatusCode == 502 || response.StatusCode == 503 || response.StatusCode == 504 {
					return errors.New("Gateway Error")
				}
				return nil
			}

		}
	}
}

func handleRequest(res http.ResponseWriter, req *http.Request) {
	var err error
	if strings.HasPrefix(req.URL.Path, "/ops/") {
		err = serveOpsPage(res, req)
	} else if reverseProxy != nil && monitor.ServerStatus.SystemReady {
		err = proxyToKubernetes(res, req)
	} else {
		err = showSiteDownPage(res)
	}

	if err != nil {
		logger.Printf("Error handling %s : %s", req.URL.Path, err)
	}

}

func serveOpsPage(res http.ResponseWriter, req *http.Request) error {
	if strings.HasPrefix(req.URL.Path, "/ops/http") {
		return nil
	} else {
		return serveLocalFile(res, req.URL.Path)
	}
}

func serveLocalFile(res http.ResponseWriter, url string) error {
	siteDownFile, err := os.Open(environment.ServerHome + "/data/web" + url)
	if err == nil {
		defer siteDownFile.Close()

		_, err = io.Copy(res, siteDownFile)
		if err != nil {
			return err
		}
	} else {
		res.WriteHeader(404)
	}
	return nil
}

func showSiteDownPage(res http.ResponseWriter) error {
	return serveLocalFile(res, "/site-down.html")
}

func proxyToKubernetes(res http.ResponseWriter, req *http.Request) error {
	reverseProxy.ServeHTTP(res, req)

	return nil
}

func generateKeys() error {
	sslCertFilePath := environment.ServerHome + "/data/ssl-cert.pem"

	_, err := os.Stat(sslCertFilePath)
	if err == nil {
		logger.Printf("Not regenerating %s", sslCertFilePath)
		return nil
	}

	logger.Printf("Generating SSL key %s", sslCertFilePath)

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Webserver"},
		},
		NotBefore: time.Now().Add(-1 * time.Hour * 48),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 2),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	//hosts := strings.Split(*host, ",")
	//for _, h := range hosts {
	//	if ip := net.ParseIP(h); ip != nil {
	//		template.IPAddresses = append(template.IPAddresses, ip)
	//	} else {
	//		template.DNSNames = append(template.DNSNames, h)
	//	}
	//}
	//
	//if *isCA {
	//	template.IsCA = true
	//	template.KeyUsage |= x509.KeyUsageCertSign
	//}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		logger.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(sslCertFilePath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(environment.ServerHome+"/data/ssl-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return err
}