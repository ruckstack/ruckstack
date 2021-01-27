package uninstall

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"os"
)

func Uninstall() error {
	packageConfig := environment.PackageConfig

	defaultValue := false
	if !ui.PromptForBoolean(fmt.Sprintf("Uninstall %s from %s", packageConfig.Name, environment.ServerHome), &defaultValue) {
		return fmt.Errorf("uninstall cancelled")
	}

	ui.Println("\nUninstalling " + packageConfig.Name + "...")

	progress := ui.StartProgressf("Shutting down server")
	if err := server.Stop(false); err != nil {
		return fmt.Errorf("error stopping server: %s", err)
	}
	progress.Stop()

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
