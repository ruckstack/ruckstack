package kubeclient

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func KubeClient() *kubernetes.Clientset {

	if !ConfigExists() {
		panic(fmt.Sprintf("%s does not exist"))
	}

	config, err := clientcmd.BuildConfigFromFlags("", KubeconfigFile())
	util.Check(err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	util.Check(err)

	return clientset
}

func ConfigExists() bool {
	_, err := os.Stat(KubeconfigFile())

	return err == nil
}

func KubeconfigFile() string {
	return util.InstallDir() + "/config/kubeconfig.yaml"
}
