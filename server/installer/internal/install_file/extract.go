package install_file

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/k3s"
	"os"
	"path/filepath"
)

type InstallFile struct {
	FilePath      string
	PackageConfig *config.PackageConfig
}

/**
Extracts the contents of this install file to the target directory.
*/
func (installFile *InstallFile) Extract(targetDir string, localConfig *config.LocalConfig) error {
	zipReader, err := zip.OpenReader(installFile.FilePath)
	if err != nil {
		return fmt.Errorf("cannot read install package: %s", err)
	}

	if err := global_util.Unzip(zipReader, targetDir); err != nil {
		return err
	}

	for file, _ := range installFile.PackageConfig.Files {
		_, err := os.Stat(filepath.Join(targetDir, file))
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("expected file %s was not installed", file)
			}
			return fmt.Errorf("error checking expected file %s: %s", file, err)
		}
		if err := installFile.PackageConfig.CheckFilePermissions(file, localConfig, targetDir); err != nil {
			return err
		}
	}

	_, err = os.Stat("/run/k3s/containerd/containerd.sock")
	if os.IsNotExist(err) {
		ui.Println("Containerd is not running. Not importing containers")
	} else {
		ui.Println("\nImporting containers...")

		imagesDir := targetDir + "/data/agent/images"
		err := filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if err := k3s.ExecCtr("images", "import", path); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
