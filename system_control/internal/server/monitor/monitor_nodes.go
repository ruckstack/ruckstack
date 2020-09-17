package monitor

import (
	"fmt"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

func watchNodes() {
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Nodes().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			newNode := newObj.(*core.Node)

			log.Printf("Monitor detected updated node: %s", newNode.Name)

			checkNode(newNode)
		},

		AddFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			log.Printf("Monitor detected added node: %s", node.Name)

			checkNode(node)

		},

		DeleteFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			log.Printf("Monitor detected deleted node %s", node.Name)

			resolveProblem(nodeIsNotReadyKey(node), fmt.Sprintf("Node %s deleted", fullName(node.ObjectMeta)))
		},
	})
	informer.Run(stopper)
}

func checkNode(node *core.Node) {
	readyCount := 0
	totalCount := 0
	for _, condition := range node.Status.Conditions {
		totalCount++
		if condition.Type == "Ready" {
			ready := condition.Status == "True"

			if ready {
				readyCount++
				resolveWarning(nodeIsNotReadyKey(node), fmt.Sprintf("Node %s is ready: %s", fullName(node.ObjectMeta), condition.Message))
			} else {
				foundWarning(nodeIsNotReadyKey(node), condition.Message)
			}
		}
	}

	noNodesReadyKey := "No nodes ready"
	if readyCount == 0 {
		foundProblem(noNodesReadyKey, fmt.Sprintf("%d of %d nodes are available", readyCount, totalCount))
	} else {
		resolveProblem(noNodesReadyKey, fmt.Sprintf("%d of %d nodes are available", readyCount, totalCount))
	}
}

func nodeIsNotReadyKey(node *core.Node) string {
	return fmt.Sprintf("Node %s is not ready", fullName(node.ObjectMeta))
}
