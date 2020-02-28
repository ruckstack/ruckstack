package monitor

import (
	"fmt"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

var (
	TRAEFIK_NOT_LISTENING = "Traefik is not listening on an IP address"
)

func watchServices() {
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Services().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newService := newObj.(*core.Service)
			log.Printf("Monitor detected updated service %s", fullName(newService.ObjectMeta))

			checkService(newService)
		},

		AddFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			log.Printf("Monitor detected added service %s", fullName(service.ObjectMeta))

			checkService(service)
		},

		DeleteFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			log.Printf("Monitor detected deleted service %s", fullName(service.ObjectMeta))

			checkService(service)
		},
	})
	informer.Run(stopper)

}

func checkService(service *core.Service) {
	fullName := fullName(service.ObjectMeta)

	if fullName == "kube-system.traefik" {
		ServerStatus.TraefikIp = service.Spec.ClusterIP

		if ServerStatus.TraefikIp == "" {
			foundProblem(TRAEFIK_NOT_LISTENING, "")
		} else {
			resolveProblem(TRAEFIK_NOT_LISTENING, fmt.Sprintf("Traefik is listening on %s", ServerStatus.TraefikIp))
		}
	}
}
