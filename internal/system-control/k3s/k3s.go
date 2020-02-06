package k3s

import (
	"github.com/ruckstack/ruckstack/internal/system-control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"log"
	"os"
	"os/exec"
	"time"
)

var k3sStartCommand *exec.Cmd

func Start() {
	log.Println("Starting K3S...")

	os.MkdirAll(util.InstallDir()+"/logs", os.FileMode(0755))

	k3sLogs, err := os.Create(util.InstallDir() + "/logs/k3s.out.logs")
	util.Check(err)
	defer k3sLogs.Close()

	k3sErrLogs, err := os.Create(util.InstallDir() + "/logs/k3s.err.logs")
	util.Check(err)
	defer k3sErrLogs.Close()

	kubecConfigFile := kubeclient.KubeconfigFile()
	k3sStartCommand = exec.Command(util.InstallDir()+"/lib/k3s", "server",
		"--bind-address", util.GetLocalConfig().BindAddress,
		"--node-external-ip", util.GetLocalConfig().BindAddress,
		"--default-local-storage-path", util.InstallDir()+"/data/local-storage",
		"--data-dir", util.InstallDir()+"/data",
		"--write-kubeconfig", kubecConfigFile,
		"--write-kubeconfig-mode", "640")
	k3sStartCommand.Stdout = k3sLogs
	k3sStartCommand.Stderr = k3sErrLogs
	err = k3sStartCommand.Start()
	util.Check(err)

	stat, err := os.Stat(kubecConfigFile)
	for stat == nil {
		time.Sleep(10 * time.Second)
		stat, err = os.Stat(kubecConfigFile)
	}

	log.Println("Starting K3S...Complete")

}

func Stop() {
	log.Println("Shutting down...")
	kill := k3sStartCommand.Process.Kill()
	if kill != nil {
		log.Printf("Error stopping k3s: %s", kill.Error())
	}
	err := k3sStartCommand.Wait()
	if err != nil {
		log.Printf("Error waiting for stopping k3s: %s", err.Error())
	}

}
