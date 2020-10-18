package install_file

import (
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func (installFile *InstallFile) Upgrade(installOptions InstallOptions) error {

	ui.Printf("Upgrading %s to version %s...", installOptions.TargetDir, installFile.PackageConfig.Version)

	serverShutdown, err := shutdownServer(installOptions.TargetDir)
	if err != nil {
		return err
	}

	localConfig, err := config.LoadLocalConfig(installOptions.TargetDir)
	if err != nil {
		return err
	}

	if err := installFile.Extract(installOptions.TargetDir, localConfig); err != nil {
		return err
	}

	ui.Println("\n\nUpgrade complete")
	ui.Println()

	if serverShutdown {
		ui.Println("Server was shut down as part of upgrade process and must be restarted")
		ui.Println()
	} else {
		ui.Println("Server was NOT auto-started as part of the upgrade process")
		ui.Println()
	}

	return nil
}

func shutdownServer(serverHome string) (bool, error) {
	serverShutdown := false
	serverPidData, err := ioutil.ReadFile(filepath.Join(serverHome, "data", "server.pid"))
	if err != nil {
		return false, err
	}

	serverPid, err := strconv.Atoi(string(serverPidData))
	if err != nil {
		return false, err
	}

	serverProcess, err := os.FindProcess(serverPid)
	err = serverProcess.Signal(syscall.Signal(0))
	if err == nil {
		ui.Printf("Found running server on PID %d", serverPid)
		ui.Println("Shutting down server...")
		ui.Println()

		if err := serverProcess.Kill(); err != nil {
			return false, err
		}

		serverShutdown = true
	} else {
		ui.Printf("No running server on PID %d", serverPid)
	}

	return serverShutdown, nil
}
