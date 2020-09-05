package restart

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Service(systemService bool, serviceName string) {
	var kubeClient = kubeclient.KubeClient()

	serviceType := "Application"
	namespace := "default"
	if systemService {
		namespace = "kube-system"
		serviceType = "System"
	}

	pods, err := kubeClient.CoreV1().Pods(namespace).List(meta.ListOptions{})
	util.Check(err)
	foundPod := false
	for _, pod := range pods.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Name == serviceName {
				err := kubeClient.CoreV1().Pods(namespace).Delete(pod.Name, &meta.DeleteOptions{})
				util.Check(err)
				foundPod = true
			}
		}
	}

	if !foundPod {
		fmt.Printf("Unknown %s service %s", strings.ToLower(serviceType), serviceName)
		return
	}

	fmt.Printf("%s service %s is restarting...\n", serviceType, serviceName)
	fmt.Println("")
	fmt.Println("Restart progress can be watched with:")
	fmt.Printf("    %s/bin/%s status services --watch\n", util.InstallDir(), util.GetPackageConfig().SystemControlName)

}
