package monitor

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

var allServices = make(map[string]*core.Service)
var TraefikIp string

func watchServices() {
	//KubeClientReady.WaitFor(true)

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Services().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldService := oldObj.(*core.Service)
			newService := newObj.(*core.Service)
			log.Printf("Changed service %s %s", oldService.Name, newService.Name)

			allServices[fullName(newService.ObjectMeta)] = newService
			checkService(newService)
		},

		AddFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			log.Printf("Added service %s", service.Name)

			allServices[fullName(service.ObjectMeta)] = service
			checkService(service)
		},

		DeleteFunc: func(obj interface{}) {
			service := obj.(*core.Service)
			fullName := fullName(service.ObjectMeta)
			log.Printf("Deleted Service %s", fullName)

			delete(allServices, fullName)
			//delete(readyDeployments, fullName)
			//delete(unreadyDeployments, fullName)
		},
	})
	informer.Run(stopper)

}

func checkService(service *core.Service) {
	fullName := fullName(service.ObjectMeta)

	log.Printf("Checking Service %s", fullName)

	if fullName == "kube-system.traefik" {
		TraefikIp = service.Spec.ClusterIP
	}

	//numberReady := service.Status.AvailableReplicas
	//if numberReady == 0 {
	//	unreadyDeployments[fullName] = numberReady
	//} else {
	//	readyDeployments[fullName] = numberReady
	//}
	//
	//DeploymentsReady = len(unreadyDeployments) == 0

}
