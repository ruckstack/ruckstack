package installer

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

func Upgrade(upgradeFile string, targetDir string) {
	util.SetInstallDir(targetDir)

	zipReader, err := zip.OpenReader(upgradeFile)
	if err != nil {
		panic(err)
	}

	var packageConfig *internal.PackageConfig
	for _, zipFile := range zipReader.File {
		if zipFile.Name == ".package.config" {
			fileReader, err := zipFile.Open()
			if err != nil {
				panic(err)
			}

			decoder := yaml.NewDecoder(fileReader)
			packageConfig = &internal.PackageConfig{}
			err = decoder.Decode(packageConfig)
			if err != nil {
				panic(err)
			}
		}
	}
	if packageConfig == nil {
		panic("Invalid upgrade file: no package config found")
	}

	fmt.Printf("Upgrading %s to %s...\n", packageConfig.Name, packageConfig.Version)

	err = extract(util.InstallDir(), zipReader)
	util.Check(err)

	fmt.Println()
	imagesDir := util.InstallDir() + "/data/agent/images"
	filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		k3s.ExecCtr("images", "import", path)

		return nil
	})

	fmt.Println("\nUpgrade complete")
}
