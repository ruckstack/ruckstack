package monitor

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

var DaemonSetsReady = false
var allDaemonSets = make(map[string]*apps.DaemonSet)
var readyDaemonSets = make(map[string]int32)
var unreadyDaemonSets = make(map[string]int32)

func watchDaemonSets() {
	//KubeClientReady.WaitFor(true)

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Apps().V1().DaemonSets().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldDaemon := oldObj.(*apps.DaemonSet)
			newDaemon := newObj.(*apps.DaemonSet)
			log.Printf("Changed daemonset %s %s", oldDaemon.Name, newDaemon.Name)

			allDaemonSets[fullName(newDaemon.ObjectMeta)] = newDaemon
			checkDaemon(newDaemon)
		},

		AddFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			log.Printf("Added daemonSet %s", daemon.Name)

			allDaemonSets[fullName(daemon.ObjectMeta)] = daemon
			checkDaemon(daemon)
		},

		DeleteFunc: func(obj interface{}) {
			daemon := obj.(*apps.DaemonSet)
			fullName := fullName(daemon.ObjectMeta)
			log.Printf("Deleted DaemonSet %s", fullName)

			delete(allDaemonSets, fullName)
			delete(readyDaemonSets, fullName)
			delete(unreadyDaemonSets, fullName)
		},
	})

	informer.Run(stopper)

}

func checkDaemon(daemon *apps.DaemonSet) {
	fullName := fullName(daemon.ObjectMeta)

	log.Printf("Checking DaemonSet %s", fullName)

	numberReady := daemon.Status.NumberReady
	if numberReady == 0 {
		unreadyDaemonSets[fullName] = numberReady
		delete(readyDaemonSets, fullName)
	} else {
		readyDaemonSets[fullName] = numberReady
		delete(unreadyDaemonSets, fullName)
	}

	DaemonSetsReady = len(unreadyDaemonSets) == 0

}
