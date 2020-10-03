package test_util

import (
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	TestTempDir string
)

func init() {
	TestTempDir = filepath.Join(global_util.GetSourceRoot(), "tmp")
	ui.VPrintf("Test temp: %s\n", TestTempDir)
}

/**
Returns the given path as a sub-path of the Ruckstack "temporary" directory.
Any "*" in the path will be replaced with a random value
*/
func TempPath(pathInTmp string) string {
	pathInTmp = strings.Replace(pathInTmp, "*", strconv.Itoa(rand.Int()), 1)
	return filepath.Join(TestTempDir, pathInTmp)
}
