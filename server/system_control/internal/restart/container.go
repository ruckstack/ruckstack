package restart

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/internal/kubeclient"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Container(systemContainer bool, containerName string) error {

	var kubeClient, err = kubeclient.KubeClient(environment.ServerHome)
	if err != nil {
		return err
	}

	containerType := "Application"
	namespace := "default"
	if systemContainer {
		namespace = "kube-system"
		containerType = "System"
	}

	if err := kubeClient.CoreV1().Pods(namespace).Delete(containerName, &meta.DeleteOptions{}); err != nil {
		return err
	}

	packageConfig := environment.PackageConfig

	fmt.Printf("%s container %s is restarting...\n", containerType, containerName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --watch\n", environment.ServerHome, packageConfig.SystemControlName)

	return nil
}
