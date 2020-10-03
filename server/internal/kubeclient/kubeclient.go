package kubeclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func KubeClient(serverHome string) (*kubernetes.Clientset, error) {

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klogFlags.Set("logtostderr", "false")
	klogFlags.Set("log_file", filepath.Join(serverHome, "logs", "k3s-client.log"))
	klog.InitFlags(klogFlags)

	if !ConfigExists(serverHome) {
		panic(fmt.Sprintf("%s does not exist", KubeconfigFile(serverHome)))
	}

	config, err := clientcmd.BuildConfigFromFlags("", KubeconfigFile(serverHome))
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

func ConfigExists(serverHome string) bool {
	_, err := os.Stat(KubeconfigFile(serverHome))

	return err == nil
}

func KubeconfigFile(serverHome string) string {
	return serverHome + "/config/kubeconfig.yaml"
}
