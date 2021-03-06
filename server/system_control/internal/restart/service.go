package restart

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Service(systemService bool, serviceName string) error {
	var kubeClient = kube.Client()

	serviceType := "Application"
	namespace := "default"
	if systemService {
		namespace = "kube-system"
		serviceType = "System"
	}

	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return err
	}
	foundPod := false
	for _, pod := range pods.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Name == serviceName {
				if err := kubeClient.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, meta.DeleteOptions{}); err != nil {
					return err
				}
				foundPod = true
			}
		}
	}

	if !foundPod {
		return fmt.Errorf("Unknown %s service %s", strings.ToLower(serviceType), serviceName)
	}

	fmt.Printf("%s service %s is restarting...\n", serviceType, serviceName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --follow\n", environment.ServerHome, environment.SystemConfig.ManagerFilename)

	return nil

}
