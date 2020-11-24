package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func checkDaemonSets(tracker *monitor.Tracker) {
	factory := informers.NewSharedInformerFactory(kube.Client(), 0)
	informer := factory.Apps().V1().DaemonSets().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newDaemon := newObj.(*apps.DaemonSet)
			tracker.Logf("Monitor detected updated daemonset %s", kube.FullName(newDaemon.ObjectMeta))

			checkDaemon(newDaemon, tracker)
		},

		AddFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			tracker.Logf("Monitor detected added daemonset %s", kube.FullName(daemon.ObjectMeta))

			checkDaemon(daemon, tracker)
		},

		DeleteFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			tracker.Logf("Monitor detected deleted daemonset %s", kube.FullName(daemon.ObjectMeta))

			tracker.ResolveComponent(kube.FullName(daemon.ObjectMeta))

		},
	})

	informer.Run(stopper)

}

func checkDaemon(daemon *apps.DaemonSet, tracker *monitor.Tracker) {
	fullName := kube.FullName(daemon.ObjectMeta)
	problemKey := fmt.Sprintf("Daemonset %s is not ready", fullName)
	numberReady := daemon.Status.NumberReady
	if numberReady == 0 {
		tracker.FoundProblem(fullName, problemKey, "No instances ready")
	} else {
		tracker.ResolveProblem(fullName, problemKey, fmt.Sprintf("Daemonset %s has at least one instance ready", fullName))
	}
}
