package kube

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var KubeconfigFile string
var client *kubernetes.Clientset

func init() {
	KubeconfigFile = environment.ServerHome + "/config/kubeconfig-admin.yaml"
}

func Client() *kubernetes.Clientset {

	if client == nil {
		var spinner ui.UiSpinner
		for true {
			_, err := os.Stat(KubeconfigFile)

			if err == nil {
				ui.VPrintf("%s was found", KubeconfigFile)
				break
			} else {
				if os.IsNotExist(err) {
					ui.VPrintf("%s does not exist yet", KubeconfigFile)
					spinner = ui.StartProgressf("Waiting for client connection details", KubeconfigFile)
					time.Sleep(time.Second * 5)
				} else {
					ui.Fatalf("cannot open %s: %s", KubeconfigFile, err)
				}
			}
		}
		if spinner != nil {
			spinner.Stop()
		}

		config, err := clientcmd.BuildConfigFromFlags("", KubeconfigFile)
		if err != nil {
			ui.Fatalf("cannot build kubernetes client config: %s", err)
		}

		client, err = kubernetes.NewForConfig(config)
		if err != nil {
			ui.Fatalf("cannot create kubernetes client: %s", err)
		}
	}

	return client
}

func FullName(obj metav1.ObjectMeta) string {
	if obj.Namespace == "" {
		return obj.Name
	}
	return obj.Namespace + "." + obj.Name
}
