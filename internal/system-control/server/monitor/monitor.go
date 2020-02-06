package monitor

import (
	"github.com/ruckstack/ruckstack/internal/system-control/files"
	"github.com/ruckstack/ruckstack/internal/system-control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"
)

var kubeClient *kubernetes.Clientset

func StartMonitor() {
	log.Println("Starting monitor...")

	for !kubeclient.ConfigExists() {
		log.Printf("Waiting for kubernetes config...")

		time.Sleep(10 * time.Second)
	}
	util.Check(files.CheckFilePermissions(util.InstallDir(), "config/kubeconfig.yaml"))

	kubeClient = kubeclient.KubeClient()
	go watchKubernetes()

	go watchNodes()
	go watchDaemonSets()
	go watchDeployments()
	go watchServices()
	go watchOverall()

	log.Println("Starting monitor...Complete")

}

func fullName(obj metav1.ObjectMeta) string {
	if obj.Namespace == "" {
		return obj.Name
	}
	return obj.Namespace + "." + obj.Name
}
