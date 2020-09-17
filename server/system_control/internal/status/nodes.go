package status

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/kubeclient"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func ShowNodeStatus(watch bool) error {
	packageConfig, err := common2.GetPackageConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Nodes in %s Cluster\n", packageConfig.Name)
	fmt.Println("----------------------------------------------------")

	kubeClient, err := kubeclient.KubeClient()
	if err != nil {
		return err
	}

	list, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	seenNodes := map[string]bool{}
	for _, node := range list.Items {
		displayNodeStatus(&node)
		seenNodes[node.Name] = true
	}

	if watch {
		fmt.Println("\nWatching for changes (ctrl-c to exit)...")
		factory := informers.NewSharedInformerFactory(kubeClient, 0)
		informer := factory.Core().V1().Nodes().Informer()
		stopper := make(chan struct{})
		defer close(stopper)

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				newNode := newObj.(*core.Node)

				displayNodeStatus(newNode)
			},

			AddFunc: func(obj interface{}) {
				node := obj.(*core.Node)

				if !seenNodes[node.Name] {
					fmt.Printf("%s (%s) has joined the cluster\n", node.Name, getNodeIp(node))

					seenNodes[node.Name] = true
					displayNodeStatus(node)
				}
			},

			DeleteFunc: func(obj interface{}) {
				node := obj.(*core.Node)

				fmt.Printf("%s (%s) has been removed from the cluster\n", node.Name, getNodeIp(node))
				seenNodes[node.Name] = false
			},
		})
		informer.Run(stopper)
	}

	return nil
}

func displayNodeStatus(node *v1.Node) {
	var nodeProblems []string
	conditionString := "UNKNOWN"

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				conditionString = "Healthy"
			} else if condition.Status == v1.ConditionFalse {
				conditionString = "ERROR"
			}
		} else {
			if condition.Status == v1.ConditionTrue {
				nodeProblems = append(nodeProblems, condition.Message)
			}
		}
	}

	nodeIp := getNodeIp(node)

	fmt.Printf("%s (%s): %s\n", node.Name, nodeIp, conditionString)
	if node.Labels["node-role.kubernetes.io/master"] == "true" {
		fmt.Println("    Primary node")
	}

	for _, problem := range nodeProblems {
		fmt.Println("    " + problem)
	}
}

func getNodeIp(node *core.Node) string {
	nodeIp := "UNKNOWN IP"
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeExternalIP {
			nodeIp = address.Address
		}
	}
	return nodeIp
}
