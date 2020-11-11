package uninstall

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Uninstall() error {
	packageConfig := environment.PackageConfig

	ui := bufio.NewScanner(os.Stdin)
	fmt.Printf("Uninstall %s from %s? [y|n] ", packageConfig.Name, environment.ServerHome)
	ui.Scan()
	if ui.Text() != "y" {
		return fmt.Errorf("cancelling install")
	}

	fmt.Println("\nUninstalling " + packageConfig.Name + "...")

	fmt.Println("Killing processes...")
	psOut := bytes.NewBufferString("")

	command := exec.Command("pgrep", "-f", environment.ServerHome)
	command.Dir = environment.ServerHome
	command.Stdout = psOut
	command.Stderr = psOut
	if err := command.Run(); err != nil {
		if err.Error() == "exit status 1" {
			//nothing matched, that is ok
		} else {
			fmt.Println(psOut.String())
			return err
		}
	}

	currentPid := os.Getpid()

	for _, pidString := range strings.Split(psOut.String(), "\n") {
		if pidString == "" {
			continue
		}
		pid, err := strconv.Atoi(pidString)
		if err != nil {
			return err
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		if process.Pid != currentPid {
			if err := process.Kill(); err != nil {
				return err
			}
		}
	}

	fmt.Println("Removing network configuration...")
	util.ExecBash("iptables-save | grep -v KUBE- | grep -v CNI- | iptables-restore")

	fmt.Println("Deleting files...")
	warn(os.RemoveAll("/etc/rancher"))
	warn(os.RemoveAll("/var/lib/rancher"))
	warn(os.RemoveAll("~/.kube"))
	warn(os.RemoveAll("/root/.kube"))
	warn(os.RemoveAll("~/.rancher"))
	warn(os.RemoveAll("/root/.rancher"))
	warn(os.RemoveAll(environment.ServerHome))

	fmt.Println("\nUninstall complete")

	return nil
}

func warn(err error) {
	if err != nil {
		fmt.Println("Warning: " + err.Error())
	}
}
