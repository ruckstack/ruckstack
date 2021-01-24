package server

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/containerd"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
)

func Stop(force bool) error {
	ctx := context.Background()

	serverProc, err := util.GetProcessFromFile(serverPidPath)
	if err != nil {
		return err
	}

	isThisProcess := serverProc != nil && serverProc.Pid == int32(os.Getpid())

	if serverProc != nil && !isThisProcess {
		err = serverProc.SendSignal(syscall.Signal(0))
		if err == nil {
			if !force {
				ui.Printf("Sending SIGTERM to %d. System will shut down in a few minutes, follow progress in the server logs", serverProc.Pid)
				if err := serverProc.TerminateWithContext(ctx); err != nil {
					ui.Fatalf("error sending SIGTERM signal: %s", err)
				}
				os.Exit(0)
			}
		}
	}

	if err := k3s.Stop(ctx); err != nil {
		return err
	}

	if err := containerd.KillProcesses(ctx); err != nil {
		return err
	}

	if serverProc != nil {
		if isThisProcess {
			ui.VPrintf("Overall server process already in process of stopping")
		} else {
			var waitGroup sync.WaitGroup
			util.ShutdownProcess(serverProc, 0, &waitGroup, ctx)
			waitGroup.Wait()
		}
	}

	if err := unmount(ctx); err != nil {
		ui.Printf("error unmounting directories: %s", err)
	}

	if err := removeNetworks(ctx); err != nil {
		ui.Printf("error removing custom networks: %s", err)
	}

	if err := os.RemoveAll("/var/lib/cni/"); err != nil {
		ui.Printf("error removing cni directory: %s", err)
	}

	if err := os.Remove(serverPidPath); err != nil {
		ui.Printf("error removing %s: %s", serverPidPath, err)
	}

	return nil
}

func removeNetworks(ctx context.Context) error {
	ui.Printf("Removing custom networks...")
	defer ui.Printf("Removing custom networks...DONE")

	//# Remove CNI namespaces
	//ip netns show 2>/dev/null | grep cni- | xargs -r -t -n 1 ip netns delete
	out, err := exec.Command("ip", "netns", "show").Output()
	if err != nil {
		return err
	}
	for _, iface := range strings.Split(string(out), "\n") {
		splitInterface := strings.Split(iface, " ")
		if len(splitInterface) < 2 {
			continue
		}

		name := strings.Trim(splitInterface[0], " ")
		if strings.Contains(name, "cni-") {
			ui.VPrintf("Deleting netns %s", name)

			_, err := exec.Command("ip", "netns", "delete", name).Output()
			if err != nil {
				ui.Printf("error deleting netns %s: %s", name, err)
			}

		}

	}

	//Delete network interface(s) that match 'master cni0'
	out, err = exec.Command("ip", "link", "show").Output()
	if err != nil {
		return err
	}
	for _, iface := range strings.Split(string(out), "\n") {
		splitInterface := strings.Split(iface, ":")
		if len(splitInterface) < 3 {
			continue
		}
		if strings.Contains(splitInterface[2], "master cni0") {
			name := strings.Trim(splitInterface[1], " ")
			name = regexp.MustCompile("@.*").ReplaceAllString(name, "")

			ui.VPrintf("Removing network %s", name)
			_, err := exec.Command("ip", "link", "delete", name).Output()
			if err != nil {
				ui.Printf("error deleting link %s: %s", name, err)
			}
		}
	}

	//remove known networks
	for _, name := range []string{"cni0", "flannel.1"} {
		ui.VPrintf("Removing network %s", name)
		_, err := exec.Command("ip", "link", "delete", name).Output()
		if err != nil && err.Error() != "exit status 1" {
			//exit status 1 == network not exists
			ui.Printf("error deleting network %s: %s", name, err)
		}
	}

	return nil
}

func unmount(context.Context) error {
	ui.Printf("Unmounting directories...")
	defer ui.Printf("Unmounting directories...DONE")

	mountContent, err := ioutil.ReadFile("/proc/self/mounts")
	if err != nil {
		return fmt.Errorf("error reading mount information: %s", err)
	}

	dirsToUnmount := []string{
		"/run/k3s/",
		"/var/lib/rancher/k3s",
		"/var/lib/kubelet/pods",
		filepath.Join(environment.ServerHome, "/data/kubelet/pods"),
		"/run/netns/cni-",
	}
	for _, mountLine := range strings.Split(string(mountContent), "\n") {
		splitLine := strings.Split(mountLine, " ")
		if len(splitLine) < 3 {
			continue
		}

		mountDir := splitLine[1]
		unmount := false
		for _, check := range dirsToUnmount {
			if strings.HasPrefix(mountDir, check) {
				unmount = true
				break
			}
		}
		if unmount {
			ui.VPrintf("Unmounting %s", mountDir)
			if err = syscall.Unmount(mountDir, syscall.MNT_DETACH); err != nil {
				ui.Printf("Cannot unmount %s", mountDir)
			} else {
				if err := os.RemoveAll(mountDir); err != nil {
					ui.Printf("Cannot remove %s", mountDir)
				}
			}
		}
	}

	return nil
}
