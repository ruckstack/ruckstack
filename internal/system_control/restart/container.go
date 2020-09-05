package restart

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Container(systemContainer bool, containerName string) {

	var kubeClient = kubeclient.KubeClient()

	containerType := "Application"
	namespace := "default"
	if systemContainer {
		namespace = "kube-system"
		containerType = "System"
	}

	err := kubeClient.CoreV1().Pods(namespace).Delete(containerName, &meta.DeleteOptions{})
	util.Check(err)

	fmt.Printf("%s container %s is restarting...\n", containerType, containerName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --watch\n", util.InstallDir(), util.GetPackageConfig().SystemControlName)

}
