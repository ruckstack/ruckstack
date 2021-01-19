package proxy

import (
	"context"
	"fmt"
	"github.com/inetaf/tcpproxy"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger
var proxies = map[string]tcpproxy.Proxy{}

func Start(ctx context.Context) error {

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "proxy.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening proxy.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	logger.Println("Starting service proxies")

	ui.Println("Starting service proxies...")

	go func() {
		select {
		case <-ctx.Done():
			logger.Println("Stopping service proxies...")
			logger.Println("Stopping proxies...DONE")
		}
	}()

	go watchServices(ctx)

	return nil
}

func watchServices(ctx context.Context) {
	kubeClient := kube.Client()

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Services().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			checkService(obj.(*core.Service))
		},

		DeleteFunc: func(obj interface{}) {
			delService := obj.(*core.Service)

			externalPort := getExternalPort(delService, 0)
			if externalPort > 0 {
				key := getFromAddress(externalPort)
				existingProxy, exists := proxies[key]
				if exists {
					log.Printf("Service %s closed. Removing proxy %s", delService.Name, key)
					if err := existingProxy.Close(); err != nil {
						log.Printf("Error closing proxy: %s", err)
					}
				}
			}
		},
	})
	informer.Run(ctx.Done())

}

func getExternalPort(service *core.Service, servicePort int32) int {
	for _, openPort := range environment.SystemConfig.Proxy {
		if openPort.ServiceName == service.Name && int32(openPort.ServicePort) == servicePort {
			return openPort.Port
		}
	}
	return -1
}

func checkService(service *core.Service) {
	for _, servicePort := range service.Spec.Ports {
		externalPort := getExternalPort(service, servicePort.Port)
		if externalPort > 0 {
			fromAddress := getFromAddress(externalPort)
			toAddress := fmt.Sprintf("%s:%d", service.Spec.ClusterIP, servicePort.Port)

			existingProxy, exists := proxies[fromAddress]
			if exists {
				logger.Printf("Replacing %s proxy %s -> %s", service.Name, fromAddress, toAddress)
				if err := existingProxy.Close(); err != nil {
					logger.Printf("Error closing existing proxy: %s", err)
				}
			} else {
				logger.Printf("Adding %s proxy %s -> %s", service.Name, fromAddress, toAddress)
			}

			var newProxy tcpproxy.Proxy
			newProxy.AddRoute(fromAddress, tcpproxy.To(toAddress))
			if err := newProxy.Start(); err != nil {
				logger.Printf("Error starting proxy: %s", err)
			}
		}
	}
}

func getFromAddress(externalPort int) string {
	return fmt.Sprintf("%s:%d", environment.LocalConfig.BindAddress, externalPort)
}
