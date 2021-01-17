package install_file

import (
	"archive/zip"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
)

func Parse(installPackagePath string) (*InstallFile, error) {
	installFile := InstallFile{
		FilePath: installPackagePath,
	}

	var err error
	zipReader, err := zip.OpenReader(installPackagePath)
	if err != nil {
		ui.Fatalf("cannot read install package: %s", err)
	}

	for _, zipFile := range zipReader.File {
		if zipFile.Name == ".package.config" {
			fileReader, err := zipFile.Open()
			if err != nil {
				ui.Fatalf("error reading package.config: %s, ", err)
			}

			installFile.PackageConfig, err = config.ReadPackageConfig(fileReader)
		} else if zipFile.Name == "config/system.config" {
			fileReader, err := zipFile.Open()
			if err != nil {
				ui.Fatalf("error reading system.config: %s, ", err)
			}

			installFile.SystemConfig, err = config.ReadSystemConfig(fileReader)

		}
	}

	return &installFile, nil
}
