package kubeclient

import (
	"flag"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var KubeconfigFile string

func init() {
	KubeconfigFile = environment.ServerHome + "/config/kubeconfig.yaml"
}

func KubeClient(serverHome string) (*kubernetes.Clientset, error) {

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klogFlags.Set("logtostderr", "false")
	klogFlags.Set("log_file", filepath.Join(serverHome, "logs", "k3s-client.log"))
	klog.InitFlags(klogFlags)

	if !ConfigExists() {
		panic(fmt.Sprintf("%s does not exist", KubeconfigFile))
	}

	config, err := clientcmd.BuildConfigFromFlags("", KubeconfigFile)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func ConfigExists() bool {
	_, err := os.Stat(KubeconfigFile)

	return err == nil
}
