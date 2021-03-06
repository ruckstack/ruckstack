package environment

import (
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	tempDirPath string
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	if global_util.IsRunningTests() {
		tempDirPath = filepath.Join(global_util.GetSourceRoot(), "tmp")
	} else {
		tempDirPath = os.TempDir()

	}

	tempDirPath = filepath.Join(tempDirPath, "ruckstack-installer")

	ui.VPrintf("Temp dir: %s\n", tempDirPath)
}

/**
Returns the given path as a sub-path of the Ruckstack "temporary" directory.
Any "*" in the path will be replaced with a random value
*/
func TempPath(pathInTmp string) string {
	pathInTmp = strings.Replace(pathInTmp, "*", strconv.Itoa(rand.Int()), 1)
	return filepath.Join(tempDirPath, pathInTmp)
}
