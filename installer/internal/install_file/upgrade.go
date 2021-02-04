package install_file

import (
	"context"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/shirou/gopsutil/v3/process"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
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

	originalPackageConfig, err := config.LoadPackageConfig(installOptions.TargetDir)
	if err != nil {
		return err
	}

	if err := originalPackageConfig.SaveBackup(installOptions.TargetDir); err != nil {
		return err
	}

	if err := installFile.Extract(installOptions.TargetDir, localConfig); err != nil {
		return err
	}

	newPackageConfig, err := config.LoadPackageConfig(installOptions.TargetDir)
	if err != nil {
		return err
	}

	for originalFile, _ := range originalPackageConfig.Files {
		keepFile := false
		for currentFile, _ := range newPackageConfig.Files {
			if currentFile == originalFile {
				keepFile = true
				break
			}
		}

		if keepFile {
			ui.VPrintf("Keeping %s: still a packaged file", originalFile)
		} else {
			ui.VPrintf("Deleting %s: outdated file", originalFile)
			err := os.Remove(filepath.Join(installOptions.TargetDir, originalFile))
			if err != nil {
				ui.Printf("Cannot delete %s: %s", originalFile, err)
			}
		}

	}

	ui.Println("\n\nUpgrade complete")
	ui.Println()

	if serverShutdown {
		ui.Println("Server was shut down as part of upgrade process and must be restarted")
		ui.Println()
	} else {
		ui.Println("Server was NOT started as part of the upgrade process")
		ui.Println()
	}

	return nil
}

func shutdownServer(serverHome string) (bool, error) {
	serverShutdown := false
	serverPidData, err := ioutil.ReadFile(filepath.Join(serverHome, "data", "server.pid"))
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	serverPid, err := strconv.Atoi(string(serverPidData))
	if err != nil {
		return false, err
	}

	serverProcess, err := process.NewProcess(int32(serverPid))
	if err == nil {
		ui.Printf("Found running server (PID %d)", serverPid)
		defer ui.StartProgressf("Shutting down server").Stop()

		var waitGroup sync.WaitGroup
		global_util.ShutdownProcess(serverProcess, 15*time.Minute, &waitGroup, context.Background())
		waitGroup.Wait()
	} else {
		ui.Printf("No running server on PID %d", serverPid)
	}

	return serverShutdown, nil
}
