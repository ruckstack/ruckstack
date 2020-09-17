package environment

import (
	"github.com/ruckstack/ruckstack/common/config"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	installDir    string
	packageConfig *config.PackageConfig
	systemConfig  *config.SystemConfig
	localConfig   *config.LocalConfig
)

func InstallDir() string {
	if installDir == "" {

		installDir = os.Getenv("RUCKSTACK_HOME")

		if installDir == "" {
			ex, exErr := os.Executable()
			if exErr != nil {
				panic(exErr)
			}
			exPath := filepath.Dir(ex)
			installDir = filepath.Dir(exPath)
		}
	}
	return installDir
}

func SetInstallDir(newInstallDir string) {
	installDir = newInstallDir
}

func SetPackageConfig(passedPackageConfig *config.PackageConfig) {
	packageConfig = passedPackageConfig
}

func GetPackageConfig() (*config.PackageConfig, error) {
	if packageConfig != nil {
		return packageConfig, nil
	}

	file, err := os.OpenFile(InstallDir()+"/.package.config", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	packageConfig = new(config.PackageConfig)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(packageConfig)
	if err != nil {
		return nil, err
	}

	return packageConfig, nil
}

func GetSystemConfig() (*config.SystemConfig, error) {
	if systemConfig != nil {
		return systemConfig, nil
	}

	file, err := os.Open(InstallDir() + "/config/system.config")
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	systemConfig = new(config.SystemConfig)
	err = decoder.Decode(systemConfig)
	if err != nil {
		return nil, err
	}

	return systemConfig, nil
}

func SetSystemConfig(passedSystemConfig *config.SystemConfig) {
	systemConfig = passedSystemConfig
}

func GetLocalConfig() (*config.LocalConfig, error) {
	if localConfig != nil {
		return localConfig, nil
	}

	file, err := os.Open(InstallDir() + "/config/local.config")
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	localConfig = new(config.LocalConfig)
	if err := decoder.Decode(localConfig); err != nil {
		return nil, err
	}

	return localConfig, nil
}

func SetLocalConfig(passedLocalConfig *config.LocalConfig) {
	localConfig = passedLocalConfig
}
