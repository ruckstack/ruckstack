package common

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	installDir    string
	packageConfig *PackageConfig
	systemConfig  *SystemConfig
	localConfig   *LocalConfig

	ruckstackHome string
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

type PackageConfig struct {
	Id                string
	Name              string
	Version           string
	BuildTime         int64  `yaml:"buildTime"`
	SystemControlName string `yaml:"systemControlName"`

	FilePermissions map[string]InstalledFileConfig `yaml:"filePermissions"`
	Files           map[string]string
}

type SystemConfig struct {
}

type LocalConfig struct {
	AdminGroup  string `yaml:"adminGroup"`
	BindAddress string `yaml:"bindAddress"`
	Join        struct {
		Server string `yaml:"server"`
		Token  string `yaml:"token"`
	} `yaml:"join"`
}

type InstalledFileConfig struct {
	AdminGroupReadable bool `yaml:"adminGroupReadable"`
	AdminGroupWritable bool `yaml:"adminGroupWritable"`
	Executable         bool
}

type AddNodeToken struct {
	Token      string `yaml:"token"`
	Server     string `yaml:"server"`
	KubeConfig string `yaml:"kubeConfig"`
}

func SetPackageConfig(passedPackageConfig *PackageConfig) {
	packageConfig = passedPackageConfig
}

func GetPackageConfig() (*PackageConfig, error) {
	if packageConfig != nil {
		return packageConfig, nil
	}

	file, err := os.OpenFile(InstallDir()+"/.package.config", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	packageConfig = new(PackageConfig)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(packageConfig)
	if err != nil {
		return nil, err
	}

	return packageConfig, nil
}

func GetSystemConfig() (*SystemConfig, error) {
	if systemConfig != nil {
		return systemConfig, nil
	}

	file, err := os.Open(InstallDir() + "/config/system.config")
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	systemConfig = new(SystemConfig)
	err = decoder.Decode(systemConfig)
	if err != nil {
		return nil, err
	}

	return systemConfig, nil
}

func SetSystemConfig(passedSystemConfig *SystemConfig) {
	systemConfig = passedSystemConfig
}

func GetLocalConfig() (*LocalConfig, error) {
	if localConfig != nil {
		return localConfig, nil
	}

	file, err := os.Open(InstallDir() + "/config/local.config")
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	localConfig = new(LocalConfig)
	if err := decoder.Decode(localConfig); err != nil {
		return nil, err
	}

	return localConfig, nil
}

func SetLocalConfig(passedLocalConfig *LocalConfig) {
	localConfig = passedLocalConfig
}

func GetRuckstackHome() string {
	if ruckstackHome != "" {
		return ruckstackHome
	}

	defaultHome := "/ruckstack"

	ruckstackHome = defaultHome
	_, err := os.Stat(ruckstackHome)
	if err == nil {
		ui.VPrintf("Ruckstack home: %s\n", ruckstackHome)
		return ruckstackHome
	}

	//No /ruckstack directory. Figure out the home directory
	executable, err := os.Executable()
	if err != nil {
		ui.Printf("Cannot determine executable. Using default home directory. Error: %s\n", err)
		return defaultHome
	}
	if executable == "ruckstack" {
		ruckstackHome = filepath.Dir(executable)
	} else {
		ruckstackHome, err = os.Getwd()

		if err != nil {
			ui.Printf("Cannot determine working directory. Using default home directory. Error: %s\n", err)
			return defaultHome
		}
	}

	for ruckstackHome != "/" {
		if _, err := os.Stat(filepath.Join(ruckstackHome, "LICENSE")); os.IsNotExist(err) {
			ruckstackHome = filepath.Dir(ruckstackHome)
			continue
		}
		break
	}

	if ruckstackHome == "/" {
		ui.VPrintf("Cannot determine Ruckstack home. Using default")
		ruckstackHome = defaultHome
	}

	ui.VPrintf("Ruckstack home: %s\n", ruckstackHome)
	return ruckstackHome

}

/**
Returns the given path as a sub-path of the Ruckstack "tmp" dir
*/
func TempPath(pathInTmp ...string) string {
	return filepath.Join(append([]string{GetRuckstackHome(), "tmp"}, pathInTmp...)...)
}
