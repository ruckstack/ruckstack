package environment

import (
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	ServerHome string
	OtherHome  string
	tempDir    string

	CurrentUser     *user.User
	IsRunningAsRoot bool

	PackageConfig *config.PackageConfig
	ClusterConfig *config.ClusterConfig
	LocalConfig   *config.LocalConfig
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

	for i, val := range os.Args {
		if val == "--server-home" && i < (len(os.Args)-1) {
			ServerHome = os.Args[i+1] + "/"
		}
	}

	if ServerHome == "" {
		if global_util.IsRunningTests() {
			ServerHome = global_util.GetSourceRoot() + "/tmp/test-installer/extracted"
		} else {
			executable, err := os.Executable()
			if err != nil {
				ui.Fatalf("Cannot determine executable: %s", err)
			}

			exPath := filepath.Dir(executable)
			ServerHome = filepath.Dir(exPath)
		}
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
		tempDir = ServerHome + "/tmp"
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ui.Fatalf("Cannot create temp dir: %s", err)
	}
	ui.VPrintf("Temp dir: %s", tempDir)

	if global_util.IsRunningTests() {
		currentUser, _ := user.Current()
		currentUserGroup, _ := user.LookupGroupId(currentUser.Gid)

		PackageConfig = &config.PackageConfig{
			Id:                "test-config",
			Name:              "Test Package",
			Version:           "0.1",
			BuildTime:         0,
			SystemControlName: "system-control",
			FilePermissions:   map[string]config.PackagedFileConfig{},
			Files:             map[string]string{},
		}

		LocalConfig = &config.LocalConfig{
			AdminGroup:  currentUserGroup.Name,
			BindAddress: "0.0.0.0",
		}
	} else {
		PackageConfig, err = config.LoadPackageConfig(ServerHome)
		if err != nil {
			ui.Fatal(err)
		}

		LocalConfig, err = config.LoadLocalConfig(ServerHome)
		if err != nil {
			ui.Fatal(err)
		}

		ClusterConfig, err = config.LoadClusterConfig(ServerHome)
		if err != nil {
			ui.Fatal(err)
		}
	}

}

/**
Returns the given path as a sub-path of the "temporary" directory.
Any "*" in the path will be replaced with a random value
*/
func TempPath(pathInTmp string) string {
	pathInTmp = strings.Replace(pathInTmp, "*", strconv.Itoa(rand.Int()), 1)
	return filepath.Join(tempDir, pathInTmp)
}
