package environment

import (
	"github.com/ruckstack/ruckstack/common/pkg/config"
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"math/rand"
	"os"
	"os/user"
	"time"
)

var (
	ServerHome string
	tempDir    string

	CurrentUser     *user.User
	IsRunningAsRoot bool

	PackageConfig *config.PackageConfig
	ClusterConfig *config.ClusterConfig
	LocalConfig   *config.LocalConfig
	SystemConfig  *config.SystemConfig

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
			ServerHome = os.Args[i+1] + "/"
		}
	}

	//TODO: set server home
	//if ServerHome == "" {
	//	....
	//}
	//_, err = os.Stat(ServerHome)
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		ui.Fatalf("Server home %s does not exist", ServerHome)
	//		return
	//	} else {
	//		ui.VPrintf("Error checking server home %s: %s", ServerHome, err)
	//		return
	//	}
	//}
	//ui.VPrintf("Server home: %s", ServerHome)

	tempDir = os.Getenv("RUCKSTACK_TEMP_DIR")
	if tempDir == "" {
		tempDir = ServerHome + "/tmp"
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
	//PackageConfig, err = config.LoadPackageConfig(ServerHome)
	//if err != nil {
	//	ui.Fatal(err)
	//}
	//
	//SystemConfig, err = config.LoadSystemConfig(ServerHome)
	//if err != nil {
	//	ui.Fatal(err)
	//}
	//
	//LocalConfig, err = config.LoadLocalConfig(ServerHome)
	//if err != nil {
	//	ui.Fatal(err)
	//}
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
