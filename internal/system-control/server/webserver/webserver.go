package webserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/server/monitor"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func StartWebserver() {
	generateKeys()

	log.Println("Starting webserver...")

	http.HandleFunc("/", handleRequest)

	go func() {
		err := http.ListenAndServe(":80", nil)
		util.Check(err)
	}()

	go func() {
		err := http.ListenAndServeTLS(":443",
			util.InstallDir()+"/data/ssl-cert.pem",
			util.InstallDir()+"/data/ssl-key.pem",
			nil)
		util.Check(err)
	}()

	log.Println("Starting webserver...complete")

	//traefikService, err := KubeClient().CoreV1().Services("kube-system").Get("traefik", metav1.GetOptions{})
	//Check(err)
	//
	//for _, port := range traefikService.Spec.Ports {
	//	switch port.Name {
	//	case "http":
	//		httpNodePort = port.NodePort
	//	case "https":
	//		httpsNodePort = port.NodePort
	//	}
	//}
	//
	//log.Printf("Http port %d", httpNodePort)
	//log.Printf("Https port %d", httpsNodePort)
}

func handleRequest(res http.ResponseWriter, req *http.Request) {
	log.Printf("handle request %s %s", req.Method, req.URL.String())

	if strings.HasPrefix(req.URL.Path, "/ops/") {
		serveOpsPage(res, req)
	} else if monitor.SystemReady {
		proxyToKubernetes(res, req)
	} else {
		showSiteDownPage(res, req)
	}

}

func serveOpsPage(res http.ResponseWriter, req *http.Request) {
	siteDownFile, err := os.Open(util.InstallDir() + "/data/web" + req.URL.Path)
	if err == nil {
		defer siteDownFile.Close()

		_, err = io.Copy(res, siteDownFile)
		util.Check(err)
	} else {
		res.WriteHeader(404)
	}
}

func proxyToKubernetes(res http.ResponseWriter, req *http.Request) {
	//res.WriteHeader(200)
	//res.Write([]byte("Tesitng server"))

	internalUrl, err := url.Parse(fmt.Sprintf("http://%s%s", monitor.TraefikIp, req.URL.String()))
	util.Check(err)
	log.Println("Proxying to " + internalUrl.String())

	reverseProxy := httputil.NewSingleHostReverseProxy(internalUrl)

	reverseProxy.ServeHTTP(res, req)
}

func generateKeys() {
	sslCertFilePath := util.InstallDir() + "/data/ssl-cert.pem"

	_, err := os.Stat(sslCertFilePath)
	if err == nil {
		log.Printf("Not regenerating %s", sslCertFilePath)
		return
	}

	log.Printf("Generating SSL key %s", sslCertFilePath)

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	util.Check(err)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	util.Check(err)

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
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(sslCertFilePath)
	util.Check(err)
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	util.Check(err)

	keyOut, err := os.OpenFile(util.InstallDir()+"/data/ssl-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	util.Check(err)
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	util.Check(err)
	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	util.Check(err)

}
