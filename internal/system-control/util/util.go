package util

import (
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	packageConfig *internal.PackageConfig
	systemConfig  *internal.SystemConfig
	localConfig   *internal.LocalConfig
	installDir    string
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

func SetPackageConfig(passedPackageConfig *internal.PackageConfig) {
	packageConfig = passedPackageConfig
}

func GetPackageConfig() *internal.PackageConfig {
	if packageConfig != nil {
		return packageConfig
	}

	file, err := os.OpenFile(InstallDir()+"/.package.config", os.O_RDONLY, 0)
	Check(err)
	defer file.Close()

	packageConfig = new(internal.PackageConfig)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(packageConfig)
	Check(err)

	return packageConfig
}

func GetSystemConfig() *internal.SystemConfig {
	if systemConfig != nil {
		return systemConfig
	}

	file, err := os.Open(InstallDir() + "/config/system.config")
	util.Check(err)
	decoder := yaml.NewDecoder(file)
	systemConfig = new(internal.SystemConfig)
	err = decoder.Decode(systemConfig)
	Check(err)

	return systemConfig
}

func SetSystemConfig(passedSystemConfig *internal.SystemConfig) {
	systemConfig = passedSystemConfig
}

func GetLocalConfig() *internal.LocalConfig {
	if localConfig != nil {
		return localConfig
	}

	file, err := os.Open(InstallDir() + "/config/local.config")
	util.Check(err)
	decoder := yaml.NewDecoder(file)
	localConfig = new(internal.LocalConfig)
	err = decoder.Decode(localConfig)
	Check(err)

	return localConfig
}

func SetLocalConfig(passedLocalConfig *internal.LocalConfig) {
	localConfig = passedLocalConfig
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func ExecBash(bashCommand string) {
	command := exec.Command("bash", "-c", bashCommand)
	command.Dir = InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		panic(err)
	}
}

func GetAbsoluteName(object meta.Object) string {
	return object.GetNamespace() + "/" + object.GetName()
}
