package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/files"
	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var k3sStartCommand *exec.Cmd

func Start() error {
	log.Println("Starting K3S...")

	os.MkdirAll(util.InstallDir()+"/logs", os.FileMode(0755))

	k3sLogs, err := os.OpenFile(util.InstallDir()+"/logs/k3s.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer k3sLogs.Close()

	kubecConfigFile := kubeclient.KubeconfigFile()
	localConfig, err := util.GetLocalConfig()
	if err != nil {
		return err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	var bindAddressIface string
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.To4().String() == localConfig.BindAddress {
				bindAddressIface = iface.Name
			}
		}
	}

	if bindAddressIface == "" {
		panic(fmt.Sprintf("Cannot find network interface with IP %s", localConfig.BindAddress))
	}

	k3sCommand := "server"
	if localConfig.Join.Server != "" {
		log.Printf("Joining server %s", localConfig.Join.Server)
		k3sCommand = "agent"
	}

	k3sArgs := []string{
		k3sCommand,
		"--node-external-ip", localConfig.BindAddress,
		"--data-dir", util.InstallDir() + "/data",
		"--kubelet-arg", "root-dir=" + util.InstallDir() + "/data/kubelet",
		//"--flannel-conf", util.InstallDir() + "/config/flannel.env",
		"--flannel-iface", bindAddressIface,
	}

	if localConfig.Join.Server == "" {
		k3sArgs = append(k3sArgs,
			"--bind-address", localConfig.BindAddress,
			"--no-deploy", "traefik",
			"--default-local-storage-path", util.InstallDir()+"/data/local-storage",
			"--write-kubeconfig", kubecConfigFile,
			"--write-kubeconfig-mode", "640",
		)
	} else {
		k3sArgs = append(k3sArgs,
			"--server", "https://"+localConfig.Join.Server+":6443",
			"--token", localConfig.Join.Token,
			"--node-ip", localConfig.BindAddress,
		)
	}
	k3sStartCommand = exec.Command(util.InstallDir()+"/lib/k3s", k3sArgs...)
	k3sStartCommand.Stdout = k3sLogs
	k3sStartCommand.Stderr = k3sLogs
	if err := k3sStartCommand.Start(); err != nil {
		return err
	}

	stat, err := os.Stat(kubecConfigFile)
	if err != nil {
		return err
	}

	for stat == nil {
		time.Sleep(10 * time.Second)
		stat, err = os.Stat(kubecConfigFile)
	}

	if err := files.CheckFilePermissions(util.InstallDir(), "config/kubeconfig.yaml"); err != nil {
		return err
	}

	log.Println("Starting K3S...Complete")

	return nil
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