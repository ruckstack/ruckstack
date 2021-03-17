package k3s

import (
	"context"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type ServiceWatcher struct {
	Namespace   string
	ServiceName string

	FireIpChanged func(string)
}

func WatchService(watcher ServiceWatcher, ctx context.Context) {
	kubeClient := kube.Client()

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Services().Informer()

	fullServiceName := watcher.Namespace + "." + watcher.ServiceName
	lastIp := ""

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			service := newObj.(*core.Service)
			if kube.FullName(service.ObjectMeta) != fullServiceName {
				return
			}

			if lastIp != service.Spec.ClusterIP {
				lastIp = service.Spec.ClusterIP
				if watcher.FireIpChanged != nil {
					watcher.FireIpChanged(lastIp)
				}
			}
		},

		AddFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			if kube.FullName(service.ObjectMeta) != fullServiceName {
				return
			}

			if lastIp != service.Spec.ClusterIP {
				lastIp = service.Spec.ClusterIP
				if watcher.FireIpChanged != nil {
					watcher.FireIpChanged(lastIp)
				}
			}
		},

		DeleteFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			if kube.FullName(service.ObjectMeta) != fullServiceName {
				return
			}

			lastIp = ""
			if watcher.FireIpChanged != nil {
				watcher.FireIpChanged(lastIp)
			}
		},
	})
	informer.Run(ctx.Done())
}
