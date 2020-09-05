package kubeclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func KubeClient() *kubernetes.Clientset {

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klogFlags.Set("logtostderr", "false")
	klogFlags.Set("log_file", filepath.Join(util.InstallDir(), "logs", "k3s-client.log"))
	klog.InitFlags(klogFlags)

	if !ConfigExists() {
		panic(fmt.Sprintf("%s does not exist", KubeconfigFile()))
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
