package monitor

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

var DeploymentsReady = false
var allDeployments = make(map[string]*apps.Deployment)
var readyDeployments = make(map[string]int32)
var unreadyDeployments = make(map[string]int32)

func watchDeployments() {
	//KubeClientReady.WaitFor(true)

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldDeployment := oldObj.(*apps.Deployment)
			newDeployment := newObj.(*apps.Deployment)
			log.Printf("Changed deployment %s %s", oldDeployment.Name, newDeployment.Name)

			allDeployments[fullName(newDeployment.ObjectMeta)] = newDeployment
			checkDeployment(newDeployment)
		},

		AddFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			log.Printf("Added deployment %s", deployment.Name)

			allDeployments[fullName(deployment.ObjectMeta)] = deployment
			checkDeployment(deployment)
		},

		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			fullName := fullName(deployment.ObjectMeta)
			log.Printf("Deleted Deployment %s", fullName)

			delete(allDeployments, fullName)
			delete(readyDeployments, fullName)
			delete(unreadyDeployments, fullName)
		},
	})
	informer.Run(stopper)

}

func checkDeployment(deployment *apps.Deployment) {
	fullName := fullName(deployment.ObjectMeta)

	log.Printf("Checking Deployment %s", fullName)

	numberReady := deployment.Status.AvailableReplicas
	if numberReady == 0 {
		unreadyDeployments[fullName] = numberReady
		delete(readyDeployments, fullName)
	} else {
		readyDeployments[fullName] = numberReady
		delete(unreadyDeployments, fullName)
	}

	DeploymentsReady = len(unreadyDeployments) == 0

}
