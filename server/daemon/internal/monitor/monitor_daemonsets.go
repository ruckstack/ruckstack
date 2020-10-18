package monitor

import (
	"fmt"
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

func watchDaemonSets() {
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().DaemonSets().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDaemon := newObj.(*apps.DaemonSet)
			log.Printf("Monitor detected updated daemonset %s", fullName(newDaemon.ObjectMeta))

			checkDaemon(newDaemon)
		},

		AddFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			log.Printf("Monitor detected added daemonset %s", fullName(daemon.ObjectMeta))

			checkDaemon(daemon)
		},

		DeleteFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			log.Printf("Monitor detected deleted daemonset %s", fullName(daemon.ObjectMeta))

			resolveProblem(daemonIsNotReadyKey(daemon), fmt.Sprintf("Daemonset %s deleted", fullName(daemon.ObjectMeta)))

		},
	})

	informer.Run(stopper)

}

func checkDaemon(daemon *apps.DaemonSet) {
	numberReady := daemon.Status.NumberReady
	if numberReady == 0 {
		foundProblem(daemonIsNotReadyKey(daemon), "No instances ready")
	} else {
		resolveProblem(daemonIsNotReadyKey(daemon), fmt.Sprintf("Daemonset %s has at least one instance ready", fullName(daemon.ObjectMeta)))
	}
}

func daemonIsNotReadyKey(daemon *apps.DaemonSet) string {
	return fmt.Sprintf("Daemonset %s is not ready", fullName(daemon.ObjectMeta))
}
