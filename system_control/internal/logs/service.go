package logs

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/kubeclient"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"strings"
	"time"
)

func ShowServiceLogs(systemService bool, serviceName string, watch bool, since string, node string) error {
	client, err := kubeclient.KubeClient()
	if err != nil {
		return err
	}

	logOptions := &core.PodLogOptions{
		Follow: watch,
	}

	fmt.Print("Logs for")

	fmt.Printf(" service %s", serviceName)
	if strings.ToLower(since) != "all" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return err
		}

		sinceSeconds := int64(math.Abs(duration.Seconds()))
		logOptions.SinceSeconds = &sinceSeconds

		fmt.Printf(" since %s", time.Now().Add(time.Duration(-1*sinceSeconds)*time.Second).Format(time.RFC822))
	}

	if watch {
		fmt.Println(" (ctrl-c to exit)...")
	} else {
		fmt.Println("")
	}
	fmt.Println("-----------------------------------------")

	namespace := "default"
	if systemService {
		namespace = "kube-system"
	}
	pods, err := client.CoreV1().Pods(namespace).List(meta.ListOptions{})
	if err != nil {
		return err
	}

	foundPod := false
	for _, pod := range pods.Items {
		if node != "all" && pod.Spec.NodeName != node {
			continue
		}
		for _, owner := range pod.OwnerReferences {
			if owner.Name == serviceName {
				foundPod = true

				if err := outputLogs(namespace, pod.Name, true, logOptions, client); err != nil {
					return err
				}
			}
		}
	}

	if !foundPod {
		fmt.Printf("No containers found for service %s", serviceName)
		if node != "all" {
			fmt.Printf(" on node %s", node)
		}
		fmt.Println("")
	}

	return nil
}
