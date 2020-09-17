package files

import (
	"github.com/ruckstack/ruckstack/common"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var adminGroupGid int

func CheckFilePermissions(installPath string, filePath string) error {
	packageConfig, err := common.GetPackageConfig()
	if err != nil {
		return err
	}

	localConfig, err := common.GetLocalConfig()
	if err != nil {
		return err
	}

	fileStat, err := os.Stat(filepath.Join(installPath, filePath))
	if err != nil {
		return nil
	}

	if adminGroupGid == 0 {
		ownerGroup, err := user.LookupGroup(localConfig.AdminGroup)
		if err != nil {
			return nil
		}
		ownerGroupUid64, err := strconv.ParseInt(ownerGroup.Gid, 10, 0)
		if err != nil {
			return nil
		}
		adminGroupGid = int(ownerGroupUid64)
	}

	var foundFileConfigPath string
	var foundFileConfig common.InstalledFileConfig

	fullFilePath := filepath.Join(installPath, filePath)
	for fileConfigPath, fileConfig := range packageConfig.FilePermissions {
		if strings.HasSuffix(fileConfigPath, "/**") {
			fileConfigPathBase := installPath + "/" + strings.Replace(fileConfigPath, "/**", "/", 1)
			if fullFilePath == fileConfigPath || strings.HasPrefix(fullFilePath, fileConfigPathBase) {
				if isBetterFileMatch(fileConfigPath, foundFileConfigPath) {
					foundFileConfig = fileConfig
					foundFileConfigPath = fileConfigPath
				}
			}
		} else {
			match, err := filepath.Match(installPath+"/"+fileConfigPath, fullFilePath)
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
		foundFileConfig = common.InstalledFileConfig{
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

	err = os.Chown(fullFilePath, 0, adminGroupGid)
	if err != nil {
		return err
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
