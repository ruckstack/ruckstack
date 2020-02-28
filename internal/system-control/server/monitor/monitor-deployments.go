package monitor

import (
	"fmt"
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

func watchDeployments() {

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDeployment := newObj.(*apps.Deployment)
			log.Printf("Monitor detected updated deployment %s", fullName(newDeployment.ObjectMeta))

			checkDeployment(newDeployment)
		},

		AddFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			log.Printf("Monitor detected added deployment %s", fullName(deployment.ObjectMeta))

			checkDeployment(deployment)
		},

		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			log.Printf("Monitor detected deleted deployment %s", fullName(deployment.ObjectMeta))

			resolveProblem(deploymentIsNotReadyKey(deployment), fmt.Sprintf("Deployment %s deleted", fullName(deployment.ObjectMeta)))
		},
	})
	informer.Run(stopper)

}

func checkDeployment(deployment *apps.Deployment) {
	numberReady := deployment.Status.AvailableReplicas
	if numberReady == 0 {
		foundProblem(deploymentIsNotReadyKey(deployment), "No instances ready")
	} else {
		resolveProblem(deploymentIsNotReadyKey(deployment), fmt.Sprintf("Deployment %s has at least one instance ready", fullName(deployment.ObjectMeta)))
	}
}

func deploymentIsNotReadyKey(deployment *apps.Deployment) string {
	return fmt.Sprintf("Deployment %s is not ready", fullName(deployment.ObjectMeta))
}
