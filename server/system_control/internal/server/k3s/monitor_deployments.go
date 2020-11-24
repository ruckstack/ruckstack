package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func checkDeployments(tracker *monitor.Tracker) {
	factory := informers.NewSharedInformerFactory(kube.Client(), 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDeployment := newObj.(*apps.Deployment)
			tracker.Logf("Monitor detected updated deployment %s", kube.FullName(newDeployment.ObjectMeta))

			checkDeployment(newDeployment, tracker)
		},

		AddFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			tracker.Logf("Monitor detected added deployment %s", kube.FullName(deployment.ObjectMeta))

			checkDeployment(deployment, tracker)
		},

		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*apps.Deployment)
			tracker.Logf("Monitor detected deleted deployment %s", kube.FullName(deployment.ObjectMeta))

			tracker.ResolveComponent(kube.FullName(deployment.ObjectMeta))
		},
	})
	informer.Run(stopper)

}

func checkDeployment(deployment *apps.Deployment, tracker *monitor.Tracker) {
	numberReady := deployment.Status.AvailableReplicas
	fullName := kube.FullName(deployment.ObjectMeta)

	problemKey := fmt.Sprintf("Deployment %s is not ready", fullName)
	if numberReady == 0 {
		tracker.FoundProblem(fullName, problemKey, "No instances ready")
	} else {
		tracker.ResolveProblem(fullName, problemKey, fmt.Sprintf("Deployment %s has at least one instance ready", fullName))
	}
}
