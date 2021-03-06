package logs

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ShowJobLogs(systemJob bool, jobName string, follow bool) error {
	client := kube.Client()

	logOptions := &core.PodLogOptions{
		Follow: follow,
	}

	fmt.Print("Logs for")

	fmt.Printf(" job %s", jobName)

	if follow {
		fmt.Println(" (ctrl-c to exit)...")
	} else {
		fmt.Println("")
	}
	fmt.Println("-----------------------------------------")

	namespace := "default"
	if systemJob {
		namespace = "kube-system"
	}
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), meta.ListOptions{})
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
