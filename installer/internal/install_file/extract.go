package install_file

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
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

	return nil
}
