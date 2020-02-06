package monitor

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
)

var ReadyNodeCount = 0

var allNodes = make(map[string]*core.Node)
var readyNodes = make(map[string]string)
var unreadyNodes = make(map[string]string)

func watchNodes() {
	//KubeClientReady.WaitFor(true)

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	informer := factory.Core().V1().Nodes().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldNode := oldObj.(*core.Node)
			newNode := newObj.(*core.Node)
			log.Printf("Changed node %s %s", oldNode.Name, newNode.Name)

			allNodes[fullName(newNode.ObjectMeta)] = newNode
			checkNode(newNode)
		},

		AddFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			log.Printf("Added node %s", node.Name)

			allNodes[fullName(node.ObjectMeta)] = node

			checkNode(node)

		},

		DeleteFunc: func(obj interface{}) {
			node := obj.(*core.Node)
			log.Printf("Deleted node %s", node.Name)

			fullName := fullName(node.ObjectMeta)
			delete(allNodes, fullName)
			delete(readyNodes, fullName)
			delete(unreadyNodes, fullName)

			ReadyNodeCount = len(readyNodes)
		},
	})
	informer.Run(stopper)
}

func checkNode(node *core.Node) {
	fullNodeName := fullName(node.ObjectMeta)

	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			ready := condition.Status == "True"

			log.Printf("Node %s ready: %t: %s", fullNodeName, ready, condition.Message)
			if ready {
				readyNodes[fullNodeName] = condition.Message
				delete(unreadyNodes, fullNodeName)
			} else {
				unreadyNodes[fullNodeName] = condition.Message
				delete(readyNodes, fullNodeName)
			}

			ReadyNodeCount = len(readyNodes)
		}
	}
}
