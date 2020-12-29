package webserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"io"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var reverseProxy *httputil.ReverseProxy

var logger *log.Logger

var sslCertFilePath = environment.ServerHome + "/data/ssl-cert.crt"
var sslKeyFilePath = environment.ServerHome + "/data/ssl-private.key"

func Start(ctx context.Context) error {

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "webserver.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	logger.Println("Starting webserver")

	ui.Println("Starting webserver...")

	http.HandleFunc("/", handleRequest)

	httpsSupported := false
	_, err = os.Stat(sslKeyFilePath)
	if err == nil {
		_, err = os.Stat(sslCertFilePath)
		if err == nil {
			go func() {
				httpsSupported = true
				logger.Println("Starting listener on port 443")
				if err := http.ListenAndServeTLS(":443",
					sslCertFilePath,
					sslKeyFilePath,
					nil); err != nil {
					e := fmt.Errorf("error starting webserver listener on port 443: %s", err)
					logger.Println(e)
					//ui.Fatal(e)
				}
			}()
		} else {
			logger.Printf("Not starting https, cannot use certificate in %s: %s", sslCertFilePath, err)
		}
	} else {
		logger.Printf("Not starting https, cannot use key in %s: %s", sslKeyFilePath, err)
	}

	go func() {
		var handler http.Handler
		if httpsSupported {
			handler = http.HandlerFunc(redirectToHttps)
		}

		logger.Println("Starting listener on port 80")
		if err := http.ListenAndServe(":80", handler); err != nil {
			e := fmt.Errorf("error starting webserver listener on port 80: %s", err)
			logger.Println(e)
			//ui.Fatal(e)
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			logger.Println("Stopping webserver...")
			logger.Println("Stopping webserver...DONE")
		}
	}()

	go watchTraefikService(ctx)

	return nil
}

var traefikIp string

func watchTraefikService(ctx context.Context) {
	kubeClient := kube.Client()

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Services().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newService := newObj.(*core.Service)

			checkTraefikService(newService)
		},

		AddFunc: func(obj interface{}) {
			checkTraefikService(obj.(*core.Service))

		},

		DeleteFunc: func(obj interface{}) {
			delService := obj.(*core.Service)

			if k3s.IsTraefik(delService) {
				traefikIp = ""
				reverseProxy = nil
				logger.Printf("Traefik service removed")
			}
		},
	})
	informer.Run(ctx.Done())

}

func checkTraefikService(service *core.Service) {
	if k3s.IsTraefik(service) {
		if traefikIp != service.Spec.ClusterIP {
			traefikIp = service.Spec.ClusterIP

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

func redirectToHttps(responseWriter http.ResponseWriter, requeset *http.Request) {
	target := "https://" + requeset.Host + requeset.URL.Path
	if len(requeset.URL.RawQuery) > 0 {
		target += "?" + requeset.URL.RawQuery
	}
	http.Redirect(responseWriter, requeset, target, http.StatusMovedPermanently)
}
