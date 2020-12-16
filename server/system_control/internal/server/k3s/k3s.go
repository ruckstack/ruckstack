package k3s

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/monitor"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

func Start(ctx context.Context) error {
	ui.Println("Starting k3s...")

	logFile := environment.ServerHome + "/logs/k3s.log"
	k3sLog, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open k3s logs: %s", err)
	}
	_, _ = k3sLog.WriteString("\n\n------ Starting k3s ---------\n")

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

			if ip.To4().String() == environment.LocalConfig.BindAddress {
				bindAddressIface = iface.Name
			}
		}
	}

	if bindAddressIface == "" {
		return fmt.Errorf("cannot find network interface with IP %s", environment.LocalConfig.BindAddress)
	}

	k3sCommand := "server"
	if environment.LocalConfig.Join.Server != "" {
		log.Printf("Joining server %s", environment.LocalConfig.Join.Server)
		k3sCommand = "agent"
	}

	k3sArgs := []string{
		k3sCommand,
		"--log", logFile,
		"--node-external-ip", environment.LocalConfig.BindAddress,
		"--data-dir", environment.ServerHome + "/data",
		"--kubelet-arg", "root-dir=" + environment.ServerHome + "/data/kubelet",
		//"--flannel-conf", util.InstallDir() + "/config/flannel.env",
		"--flannel-iface", bindAddressIface,
	}
	if ui.IsVerbose() {
		k3sArgs = append(k3sArgs, "--debug")
	}

	if environment.LocalConfig.Join.Server == "" {
		k3sArgs = append(k3sArgs,
			"--cluster-init",
			"--bind-address", environment.LocalConfig.BindAddress,
			"--no-deploy", "traefik",
			"--default-local-storage-path", environment.ServerHome+"/data/local-storage",
			"--write-kubeconfig", kube.KubeconfigFile,
			"--write-kubeconfig-mode", "640",
		)
	} else {
		k3sArgs = append(k3sArgs,
			"--server", "https://"+environment.LocalConfig.Join.Server+":6443",
			"--token", environment.LocalConfig.Join.Token,
			"--node-ip", environment.LocalConfig.BindAddress,
		)
	}

	_, _ = k3sLog.WriteString(fmt.Sprintf("Running k3s %s\n", strings.Join(k3sArgs, " ")))
	_ = k3sLog.Close()

	k3sStartCommand := exec.Command(environment.ServerHome+"/lib/k3s", k3sArgs...)
	if err := k3sStartCommand.Start(); err != nil {
		return err
	}

	monitor.Add(&monitor.Tracker{
		Name:  "Kubernetes Client",
		Check: checkKubeClient,
	})

	monitor.Add(&monitor.Tracker{
		Name:  "Managed Traefik",
		Check: checkTraefik,
	})

	monitor.Add(&monitor.Tracker{
		Name:  "Managed DaemonSets",
		Check: checkDaemonSets,
	})

	monitor.Add(&monitor.Tracker{
		Name:  "Managed Deployments",
		Check: checkDeployments,
	})

	monitor.Add(&monitor.Tracker{
		Name:  "Server Nodes",
		Check: checkNodes,
	})

	client := kube.Client()
	version, err := client.ServerVersion()
	if err != nil {
		return fmt.Errorf("error checking server: %s", err)
	}
	ui.Printf("Server version %s started", version.String())

	if err := os.Chown(kube.KubeconfigFile, 0, int(environment.LocalConfig.AdminGroupId)); err != nil {
		ui.Fatalf("Cannot set %s ownership: %s", kube.KubeconfigFile, err)
	}

	go func() {
		select {
		case <-ctx.Done():
			ui.Println("Server shutting down...")
			ui.VPrintf("Shutdown reason: %s", ctx.Err())
		}
	}()

	return nil
}
