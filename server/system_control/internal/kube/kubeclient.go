package kube

import (
	"flag"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var KubeconfigFile string
var client *kubernetes.Clientset

func init() {
	KubeconfigFile = environment.ServerHome + "/config/kubeconfig-admin.yaml"
}

func Client() *kubernetes.Clientset {

	if client == nil {
		logFile := filepath.Join(environment.ServerHome, "logs", "kubernetes.client.log")
		clientLog, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			ui.Fatalf("cannot open k3s.client.log: %s", err)
		}
		_, _ = clientLog.WriteString("\n\n------ Starting Kubernetes Client  ---------\n")

		klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
		_ = klogFlags.Set("logtostderr", "false")
		_ = klogFlags.Set("log_file", logFile)
		klog.InitFlags(klogFlags)

		for true {
			_, err := os.Stat(KubeconfigFile)

			if err == nil {
				break
			} else {
				if os.IsNotExist(err) {
					_, _ = clientLog.WriteString(fmt.Sprintf("Waiting for %s to be created...", KubeconfigFile))
					time.Sleep(time.Second * 5)
				} else {
					ui.Fatalf("cannot open %s: %s", KubeconfigFile, err)
				}
			}
		}
		_ = clientLog.Close()

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
