package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"

	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func checkNodes(tracker *monitor.Tracker) {
	factory := informers.NewSharedInformerFactory(kube.Client(), 0)
	informer := factory.Core().V1().Nodes().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newNode := newObj.(*core.Node)

			tracker.Logf("Monitor detected updated node: %s", newNode.Name)

			checkNode(newNode, tracker)
		},

		AddFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			tracker.Logf("Monitor detected added node: %s", node.Name)

			checkNode(node, tracker)

		},

		DeleteFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			tracker.Logf("Monitor detected deleted node %s", node.Name)

			tracker.ResolveComponent(kube.FullName(node.ObjectMeta))
		},
	})
	informer.Run(stopper)
}

func checkNode(node *core.Node, tracker *monitor.Tracker) {
	readyCount := 0
	totalCount := 0
	for _, condition := range node.Status.Conditions {
		totalCount++
		if condition.Type == "Ready" {
			ready := condition.Status == "True"

			fullName := kube.FullName(node.ObjectMeta)
			problemKey := fmt.Sprintf("Node %s is not ready", fullName)

			if ready {
				readyCount++
				tracker.ResolveWarning(fullName, problemKey, fmt.Sprintf("Node %s is ready: %s", fullName, condition.Message))
			} else {
				tracker.FoundWarning(fullName, problemKey, condition.Message)
			}
		}
	}

	noNodesReadyKey := "No nodes ready"
	if readyCount == 0 {
		tracker.FoundProblem("k3s", noNodesReadyKey, fmt.Sprintf("%d of %d nodes are available", readyCount, totalCount))
	} else {
		tracker.ResolveProblem("k3s", noNodesReadyKey, fmt.Sprintf("%d of %d nodes are available", readyCount, totalCount))
	}
}
