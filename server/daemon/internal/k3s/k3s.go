package k3s

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/rancher/k3s/pkg/cli/cmds"
	k3sServer "github.com/rancher/k3s/pkg/cli/server"
	"github.com/rancher/k3s/pkg/configfilearg"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
)

var (
	ServerToken string
)

var ContainerdClient *containerd.Client
var KubeClient *kubernetes.Clientset

var containerdAddress = "/run/k3s/containerd/containerd.sock"

func Start(parent context.Context) error {
	ui.Println("Starting k3s...")

	kubeconfigPath := filepath.Join(environment.ServerHome, "/config/kubeconfig.yaml")

	go func() {
		select {
		case <-parent.Done():
			if ContainerdClient != nil {
				ui.Println("Stopping images...")
				//ContainerdClient.Li.Close()
				ui.Println("Stopping images...DONE")
			}
		}
	}()

	k3sArgs := []string{
		"k3s",
		"server",
		"--log", environment.ServerHome + "/logs/k3s.log",
		"--alsologtostderr", "false",
		"--node-external-ip", environment.LocalConfig.BindAddress,
		"--data-dir", environment.ServerHome + "/data",
		"--kubelet-arg", "root-dir=" + environment.ServerHome + "/data/kubelet",
		"--flannel-iface", environment.LocalConfig.BindAddressInterface,
		"--datastore-endpoint", "https://localhost:2379",
	}

	if ui.IsVerbose() {
		k3sArgs = append(k3sArgs, "--debug")
	}

	//if localConfig.Join.Server == "" {
	k3sArgs = append(k3sArgs,
		"--no-deploy", "traefik",
		"--default-local-storage-path", environment.ServerHome+"/data/local-storage",
		"--write-kubeconfig", kubeconfigPath,
		"--write-kubeconfig-mode", "640",
	)
	//} else {
	//	k3sArgs = append(k3sArgs,
	//		"--server", "https://"+localConfig.Join.Server+":6443",
	//		"--token", localConfig.Join.Token,
	//		"--node-ip", localConfig.BindAddress,
	//	)
	//}

	go func() {
		app := cmds.NewApp()
		app.Commands = []cli.Command{
			cmds.NewServerCommand(k3sServer.Run),
		}

		commandArgs := configfilearg.MustParse(k3sArgs)

		os.Setenv("_K3S_LOG_REEXEC_", "true")
		if err := app.Run(commandArgs); err != nil {
			ui.Fatalf("error starting k3s: %s", err)
		}
	}()

	//var err error
	//for ContainerdClient == nil {
	//	ui.Printf("Connecting to containerd at %s", containerdAddress)
	//	ContainerdClient, err = containerd.New(containerdAddress)
	//	if err != nil {
	//		ui.Printf("Waiting for containerd...%s", err)
	//		time.Sleep(2 * time.Second)
	//	}
	//}
	//config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	//if err != nil {
	//	return err
	//}
	//
	//// create the clientset
	//KubeClient, err = kubernetes.NewForConfig(config)
	//if err != nil {
	//	return err
	//}

	return nil
}
