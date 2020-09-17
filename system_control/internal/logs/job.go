package logs

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/kubeclient"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ShowJobLogs(systemJob bool, jobName string, watch bool) error {
	client, err := kubeclient.KubeClient()
	if err != nil {
		return err
	}

	logOptions := &core.PodLogOptions{
		Follow: watch,
	}

	fmt.Print("Logs for")

	fmt.Printf(" job %s", jobName)

	if watch {
		fmt.Println(" (ctrl-c to exit)...")
	} else {
		fmt.Println("")
	}
	fmt.Println("-----------------------------------------")

	namespace := "default"
	if systemJob {
		namespace = "kube-system"
	}
	pods, err := client.CoreV1().Pods(namespace).List(meta.ListOptions{})
	if err != nil {
		return err
	}

	foundPod := false
	for _, pod := range pods.Items {

		for _, owner := range pod.OwnerReferences {
			if owner.Name == jobName {
				foundPod = true

				if err := outputLogs(namespace, pod.Name, true, logOptions, client); err != nil {
					return err
				}
			}
		}
	}

	if !foundPod {
		fmt.Printf("No containers found for job %s", jobName)

		fmt.Println("")
	}

	return nil
}
