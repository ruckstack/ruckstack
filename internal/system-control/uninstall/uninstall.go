package uninstall

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Uninstall() {
	packageConfig := util.GetPackageConfig()

	ui := bufio.NewScanner(os.Stdin)
	fmt.Printf("Uninstall %s from %s? [y|n] ", packageConfig.Name, util.InstallDir())
	ui.Scan()
	if ui.Text() != "y" {
		fmt.Printf("Cancelling install")
		os.Exit(1)
	}

	fmt.Println("\nUninstalling " + packageConfig.Name + "...")

	fmt.Println("Killing processes...")
	psOut := bytes.NewBufferString("")

	command := exec.Command("pgrep", "-f", util.InstallDir())
	command.Dir = util.InstallDir()
	command.Stdout = psOut
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		if err.Error() == "exit status 1" {
			//nothing matched, that is ok
		} else {
			fmt.Println(psOut.String())
			panic(err)
		}
	}

	currentPid := os.Getpid()

	for _, pidString := range strings.Split(psOut.String(), "\n") {
		if pidString == "" {
			continue
		}
		pid, err := strconv.Atoi(pidString)
		util.Check(err)
		process, err := os.FindProcess(pid)
		util.Check(err)

		if process.Pid != currentPid {
			err := process.Kill()
			util.Check(err)
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
	warn(os.RemoveAll(util.InstallDir()))

	fmt.Println("\nUninstall complete")
}

func warn(err error) {
	if err != nil {
		fmt.Println("Warning: " + err.Error())
	}
}
