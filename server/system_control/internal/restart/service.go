package restart

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kubeclient"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Service(systemService bool, serviceName string) error {
	var kubeClient, err = kubeclient.KubeClient(environment.ServerHome)
	if err != nil {
		return err
	}

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

	packageConfig := environment.PackageConfig

	fmt.Printf("%s service %s is restarting...\n", serviceType, serviceName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --watch\n", environment.ServerHome, packageConfig.SystemControlName)

	return nil

}
