package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ruckstack/ruckstack/server/system_control/internal/dev"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"strings"
)

var serviceIngresses = map[string][]string{}
var ingressDevProxies = map[string]DevModeProxy{}
var proxyConfigs = map[string]dev.ProxyConfig{}

func init() {
	Register(&RouterPlugin{})
}

func (plugin *RouterPlugin) Name() string {
	return "Traefik Proxy"
}

func (plugin *RouterPlugin) Start(router *gin.Engine, ctx context.Context) {
	proxy := &ServiceProxy{
		Namespace:   "kube-system",
		ServiceName: "traefik",
	}

	go watchIngress(ctx)
	go watchDevConfig(ctx)
	go proxy.Start(ctx)

	router.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/ops") {
			serveOpsPage(ctx)
		} else {
			if environment.ClusterConfig.DevModeEnabled {
				matchedPath := ""
				var matchedProxy *DevModeProxy
				for ingress, foundProxy := range ingressDevProxies {
					if len(ingress) > len(matchedPath) && strings.HasPrefix(ctx.Request.URL.Path, ingress) {
						matchedPath = ingress
						matchedProxy = &foundProxy
					}
				}

				if matchedProxy != nil {
					matchedProxy.RequestHandler(ctx)
					return
				}
			}

			proxy.RequestHandler(ctx)
		}
	})
}

func watchIngress(ctx context.Context) {
	client := kube.Client()

	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Networking().V1().Ingresses().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress := obj.(*networking.Ingress)
			if ingress.ObjectMeta.Namespace != "default" {
				return
			}

			saveIngressSettings(ingress)
			setupDevProxies(ctx)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ingress := newObj.(*networking.Ingress)
			if ingress.ObjectMeta.Namespace != "default" {
				return
			}

			saveIngressSettings(ingress)
			setupDevProxies(ctx)
		},
		DeleteFunc: func(obj interface{}) {
			ingress := obj.(*networking.Ingress)
			if ingress.ObjectMeta.Namespace != "default" {
				return
			}

			for _, rule := range ingress.Spec.Rules {
				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						delete(serviceIngresses, path.Backend.Service.Name)
						delete(serviceIngresses, path.Backend.Service.Name)
					}
				}
			}

			setupDevProxies(ctx)
		},
	})

	informer.Run(ctx.Done())
}

func saveIngressSettings(ingress *networking.Ingress) {
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				paths, ok := serviceIngresses[path.Backend.Service.Name]
				if !ok {
					paths = []string{}
				}
				paths = append(paths, path.Path)
				serviceIngresses[path.Backend.Service.Name] = paths
			}
		}
	}
}

func watchDevConfig(ctx context.Context) {
	client := kube.Client()

	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().ConfigMaps().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap := obj.(*core.ConfigMap)
			if kube.FullName(configMap.ObjectMeta) != "kube-system.dev-config" {
				return
			}

			configureDevMode(configMap.Data["enabled"] == "true")
			saveDevProxyConfig(configMap)
			setupDevProxies(ctx)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			configMap := newObj.(*core.ConfigMap)
			if kube.FullName(configMap.ObjectMeta) != "kube-system.dev-config" {
				return
			}

			configureDevMode(configMap.Data["enabled"] == "true")
			saveDevProxyConfig(configMap)
			setupDevProxies(ctx)
		},
		DeleteFunc: func(obj interface{}) {
			configMap := obj.(*core.ConfigMap)
			if kube.FullName(configMap.ObjectMeta) != "kube-system.dev-config" {
				return
			}

			configureDevMode(false)
			proxyConfigs = map[string]dev.ProxyConfig{}
			setupDevProxies(ctx)
		},
	})

	informer.Run(ctx.Done())
}

func saveDevProxyConfig(configMap *core.ConfigMap) {
	proxyString := configMap.Data["proxy"]
	if proxyString != "" {
		decoder := yaml.NewDecoder(strings.NewReader(proxyString))
		if err := decoder.Decode(proxyConfigs); err != nil {
			logger.Printf("cannt parse proxy config: %s", err)
		}
	}
}

func configureDevMode(enabled bool) {
	if environment.ClusterConfig.DevModeEnabled != enabled {
		environment.ClusterConfig.DevModeEnabled = enabled
		if err := dev.SaveDevMode(enabled); err != nil {
			logger.Printf("error saving dev mode: %s", err)
		}
	}
}

func setupDevProxies(ctx context.Context) {
	for service, devConfig := range proxyConfigs {
		proxy := DevModeProxy{
			targetHost: devConfig.TargetHost,
			targetPort: devConfig.TargetPort,
		}
		proxy.Start(ctx)

		ingresses := serviceIngresses[service]
		if len(ingresses) == 0 {
			logger.Printf("Dev redirect configured for service %s, but no ingresses were defined", service)
		} else {
			for _, ingress := range ingresses {
				ingressDevProxies[ingress] = proxy
			}
		}
	}
}

type RouterPlugin struct {
}
