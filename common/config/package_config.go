package config

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

type PackageConfig struct {
	Id              string
	Name            string
	Version         string
	BuildTime       int64  `yaml:"buildTime"`
	ManagerFilename string `yaml:"managerFilename"`

	FilePermissions map[string]PackagedFileConfig `yaml:"filePermissions"`
	Files           map[string]string
}

type PackagedFileConfig struct {
	AdminGroupReadable bool `yaml:"adminGroupReadable"`
	AdminGroupWritable bool `yaml:"adminGroupWritable"`
	Executable         bool
}

func ReadPackageConfig(content io.ReadCloser) (*PackageConfig, error) {
	packageConfig := new(PackageConfig)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(packageConfig); err != nil {
		return nil, fmt.Errorf("error parsing package.config: %s, ", err)
	}

	return packageConfig, nil
}

func LoadPackageConfig(serverHome string) (*PackageConfig, error) {
	file, err := os.Open(serverHome + "/.package.config")
	if err != nil {
		return nil, err
	}

	return ReadPackageConfig(file)
}

func (packageConfig *PackageConfig) Save(serverHome string, localConfig *LocalConfig) error {
	packageConfigFile, err := os.OpenFile(serverHome+"/.package.config", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	packageConfigEncoder := yaml.NewEncoder(packageConfigFile)
	if err := packageConfigEncoder.Encode(packageConfig); err != nil {
		return err
	}
	if err := packageConfig.CheckFilePermissions(".package.config", localConfig, serverHome); err != nil {
		return err
	}

	return nil
}

func (packageConfig *PackageConfig) CheckFilePermissions(filePath string, localConfig *LocalConfig, serverHome string) error {
	currentUser, err := user.Current()
	if err != nil {
		ui.Printf("Cannot determine current user: %s", err)
		return err
	}

	fileStat, err := os.Stat(filepath.Join(serverHome, filePath))
	if err != nil {
		return fmt.Errorf("Cannot check file %s: %s", filePath, err)
	}

	ownerGroup, err := user.LookupGroup(localConfig.AdminGroup)
	if err != nil {
		return err
	}
	ownerGroupUid64, err := strconv.ParseInt(ownerGroup.Gid, 10, 0)
	if err != nil {
		return err
	}

	adminGroupGid := int(ownerGroupUid64)

	var foundFileConfigPath string
	var foundFileConfig PackagedFileConfig

	fullFilePath := filepath.Join(serverHome, filePath)
	for fileConfigPath, fileConfig := range packageConfig.FilePermissions {
		if strings.HasSuffix(fileConfigPath, "/**") {
			fileConfigPathBase := serverHome + "/" + strings.Replace(fileConfigPath, "/**", "/", 1)
			if fullFilePath == fileConfigPath || strings.HasPrefix(fullFilePath, fileConfigPathBase) {
				if isBetterFileMatch(fileConfigPath, foundFileConfigPath) {
					foundFileConfig = fileConfig
					foundFileConfigPath = fileConfigPath
				}
			}
		} else {
			match, err := filepath.Match(serverHome+"/"+fileConfigPath, fullFilePath)
			if err != nil {
				return err
			}

			if match {
				if isBetterFileMatch(fileConfigPath, foundFileConfigPath) {
					foundFileConfig = fileConfig
					foundFileConfigPath = fileConfigPath
				}
			}
		}
	}

	if foundFileConfigPath == "" {
		foundFileConfig = PackagedFileConfig{
			AdminGroupReadable: false,
			AdminGroupWritable: false,
			Executable:         false,
		}
	}

	//if isIgnorePermissions(filePath) {
	//	if info.IsDir() {
	//		return filepath.SkipDir
	//	} else {
	//		return nil
	//	}
	//}

	if currentUser.Name == "root" {
		err = os.Chown(fullFilePath, 0, adminGroupGid)
		if err != nil {
			return err
		}
	} else {
		//normal code will run as root, but tests run as a random user
		ui.VPrintf("Cannot change %s owner to root, because not running as root", fullFilePath)
	}

	var mode os.FileMode
	if !foundFileConfig.AdminGroupReadable {
		if fileStat.IsDir() || foundFileConfig.Executable {
			mode = os.FileMode(0700)
		} else {
			mode = os.FileMode(0600)
		}
	} else {
		if foundFileConfig.AdminGroupWritable {
			if fileStat.IsDir() || foundFileConfig.Executable {
				mode = os.FileMode(0770)
			} else {
				mode = os.FileMode(0660)
			}
		} else {
			if fileStat.IsDir() || foundFileConfig.Executable {
				mode = os.FileMode(0750)
			} else {
				mode = os.FileMode(0640)
			}
		}
	}

	err = os.Chmod(fullFilePath, mode)
	if err != nil {
		return nil
	}

	return nil
}

func isBetterFileMatch(newFilePath string, existingFilePath string) bool {
	if existingFilePath == "" {
		return true
	}

	if strings.HasSuffix(newFilePath, "/**") && !strings.HasSuffix(existingFilePath, "/**") {
		return false
	}
	if strings.HasSuffix(newFilePath, "/*") && !strings.HasSuffix(existingFilePath, "/*") {
		return false
	}
	if len(newFilePath) > len(existingFilePath) {
		return true
	}
	return true
}
