package restart

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Container(systemContainer bool, containerName string) error {

	var kubeClient = kube.Client()

	containerType := "Application"
	namespace := "default"
	if systemContainer {
		namespace = "kube-system"
		containerType = "System"
	}

	if err := kubeClient.CoreV1().Pods(namespace).Delete(context.Background(), containerName, meta.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Printf("%s container %s is restarting...\n", containerType, containerName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --follow\n", environment.ServerHome, environment.SystemConfig.ManagerFilename)

	return nil
}
