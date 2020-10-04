package environment

import (
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
	CurrentUser *user.User

	tempDir string
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	var err error
	CurrentUser, err = user.Current()
	if err != nil {
		ui.Printf("Cannot determine current user: %s", err)
		return
	}

	tempDir = os.Getenv("RUCKSTACK_TEMP_DIR")
	if tempDir == "" {
		//find tempDir
		if global_util.IsRunningTests() {
			tempDir = global_util.GetSourceRoot() + "/tmp"
		} else {
			tempDir = filepath.Join(os.TempDir(), "ruckstack-launcher")
		}
	}
	ui.VPrintf("Ruckstack launcher temp dir: %s", tempDir)
}

/**
Returns the given path as a sub-path of the Ruckstack "temporary" directory.
Any "*" in the path will be replaced with a random value
*/
func TempPath(pathInTmp string) string {
	pathInTmp = strings.Replace(pathInTmp, "*", strconv.Itoa(rand.Int()), 1)
	return filepath.Join(tempDir, pathInTmp)
}
