package config

import (
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var (
	ServerHome string
	tempDir    string

	CurrentUser     *user.User
	IsRunningAsRoot bool

	PackageConfig *PackageConfigType
	//ClusterConfig *config.ClusterConfig
	LocalConfig *LocalConfigType
	//SystemConfig  *config.SystemConfig

	NodeName string
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	var err error
	CurrentUser, err = user.Current()
	if err != nil {
		ui.Printf("Cannot determine current user: %s", err)
		return
	}
	IsRunningAsRoot = CurrentUser.Name == "root"

	NodeName, err = os.Hostname()
	if err != nil {
		ui.VPrintf("Running on nodeName %s", NodeName)
		ui.Printf("Cannot determine hostname: %s", err)
		return
	}

	for i, val := range os.Args {
		if val == "--server-home" && i < (len(os.Args)-1) {
			ServerHome = filepath.FromSlash(os.Args[i+1] + "/")
			break
		}
	}

	if ServerHome == "" {
		executable, err := os.Executable()
		if err != nil {
			ui.Fatalf("Cannot determine executable: %s", err)
		}

		exPath := filepath.Dir(executable)
		ServerHome = filepath.Dir(exPath)
	}

	_, err = os.Stat(ServerHome)
	if err != nil {
		if os.IsNotExist(err) {
			ui.Fatalf("Server home %s does not exist", ServerHome)
			return
		} else {
			ui.VPrintf("Error checking server home %s: %s", ServerHome, err)
			return
		}
	}
	ui.VPrintf("Server home: %s", ServerHome)

	tempDir = os.Getenv("RUCKSTACK_TEMP_DIR")
	if tempDir == "" {
		tempDir = filepath.FromSlash(ServerHome + "/tmp")
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ui.Fatalf("Cannot create temp dir: %s", err)
	}
	ui.VPrintf("Temp dir: %s", tempDir)

	//if global_util.IsRunningTests() { //TODO keep this?
	//	currentUser, _ := user.Current()
	//	currentUserGroup, _ := user.LookupGroupId(currentUser.Gid)
	//
	//	PackageConfig = &config.PackageConfig{
	//		Id:              "test-config",
	//		Name:            "Test Package",
	//		Version:         "0.1",
	//		BuildTime:       0,
	//		FilePermissions: map[string]config.PackagedFileConfig{},
	//		Files:           map[string]string{},
	//	}
	//
	//	LocalConfig = &config.LocalConfig{
	//		AdminGroup:  currentUserGroup.Name,
	//		BindAddress: "127.0.0.1",
	//	}
	//
	//	SystemConfig = &config.SystemConfig{
	//		ManagerFilename: "test-control",
	//	}
	//} else {
	PackageConfig, err = LoadPackageConfig(ServerHome)
	if err != nil {
		ui.Fatal(err)
	}
	//
	//SystemConfig, err = config.LoadSystemConfig(ServerHome)
	//if err != nil {
	//	ui.Fatal(err)
	//}
	//
	LocalConfig, err = LoadLocalConfig(ServerHome)
	if err != nil {
		if os.IsNotExist(err) {
			//not set up yet
		} else {
			ui.Fatal(err)
		}
	}
	//
	//ClusterConfig, err = config.LoadClusterConfig(ServerHome)
	//if err != nil {
	//	ui.Fatal(err)
	//}
	//}

	//adminGroup, err := user.LookupGroup(LocalConfig.AdminGroup)
	//if err != nil {
	//	ui.Fatalf("Cannot find admin group %s", adminGroup)
	//	return
	//}
	//
	//LocalConfig.AdminGroupId, err = strconv.ParseInt(adminGroup.Gid, 10, 0)
	//if err != nil {
	//	ui.Fatalf("Error parsing admin group id %s", adminGroup.Gid)
	//}

}
