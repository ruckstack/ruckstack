package internal

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/k3s"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func Upgrade(upgradeFile string, targetDir string) error {

	upgradeLog, err := os.OpenFile(filepath.Join(targetDir, "logs", "upgrade-"+strconv.FormatInt(time.Now().Unix(), 10)+".log"), os.O_WRONLY|os.O_CREATE, 0644)
	defer upgradeLog.Close()

	log.SetOutput(upgradeLog)
	log.Printf("Upgrading %s from %s", targetDir, upgradeFile)

	if err != nil {
		return err
	}

	environment.SetInstallDir(targetDir)

	zipReader, err := zip.OpenReader(upgradeFile)
	if err != nil {
		panic(err)
	}

	var packageConfig *config.PackageConfig
	for _, zipFile := range zipReader.File {
		if zipFile.Name == ".package.config" {
			fileReader, err := zipFile.Open()
			if err != nil {
				panic(err)
			}

			decoder := yaml.NewDecoder(fileReader)
			packageConfig = &config.PackageConfig{}
			err = decoder.Decode(packageConfig)
			if err != nil {
				panic(err)
			}
		}
	}
	if packageConfig == nil {
		panic("Invalid upgrade file: no package config found")
	}

	userMessage("Upgrading %s to version %s...\n\n", packageConfig.Name, packageConfig.Version)

	serverShutdown := false
	serverPidData, err := ioutil.ReadFile(filepath.Join(environment.InstallDir(), "data", "server.pid"))
	if err != nil {
		return err
	}

	serverPid, err := strconv.Atoi(string(serverPidData))
	if err != nil {
		return err
	}

	serverProcess, err := os.FindProcess(serverPid)
	err = serverProcess.Signal(syscall.Signal(0))
	if err == nil {
		log.Printf("Found running server on PID %d", serverPid)
		userMessage("Shutting down server...\n")

		if err := serverProcess.Kill(); err != nil {
			return err
		}

		serverShutdown = true
	} else {
		log.Printf("No running server on PID %d", serverPid)
	}

	userMessage("Extracting files...")
	if err := global_util.Unzip(zipReader, environment.InstallDir()); err != nil {
		return err
	}

	_, err = os.Stat("/run/k3s/containerd/containerd.sock")
	if os.IsNotExist(err) {
		log.Println("Containerd is not running. Not importing containers")
	} else {
		userMessage("\nImporting containers...")

		imagesDir := environment.InstallDir() + "/data/agent/images"
		filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			fmt.Print(".")
			k3s.ExecCtr(log.Writer(), log.Writer(), "images", "import", path)

			return nil
		})
	}

	userMessage("\n\nUpgrade complete\n")

	if serverShutdown {
		userMessage("Server was shut down as part of upgrade process and must be restarted\n")
	} else {
		userMessage("Server was NOT auto-started as part of the upgrade process\n")
	}

	return nil
}

func userMessage(message string, args ...interface{}) {
	log.Printf(message, args...)
	fmt.Printf(message, args...)

}
