package environment

import (
	"fmt"
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
	RuckstackWorkDir string
	cacheRoot        string
	tempDir          string

	OutDir     string
	ProjectDir string

	PackagedK3sVersion  = "1.20.4+k3s1"
	PackagedHelmVersion = "3.4.2" //should match go.mod

)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	RuckstackWorkDir = os.Getenv("RUCKSTACK_WORK_DIR")
	if RuckstackWorkDir == "" {
		if global_util.IsRunningTests() {
			RuckstackWorkDir = global_util.GetSourceRoot() + "/tmp/work_dir"
		} else {
			userHome, err := os.UserHomeDir()
			if err != nil || userHome == "" {
				ui.Fatalf("Cannot determine home directory: %s", err)
			}

			RuckstackWorkDir = filepath.Join(userHome, ".ruckstack")
		}
	}

	ui.VPrintf("Ruckstack work directory: %s\n", RuckstackWorkDir)

	cacheRoot = filepath.Join(RuckstackWorkDir, "cache")
	if err := os.MkdirAll(cacheRoot, 0755); err != nil {
		panic(fmt.Sprintf("Cannot create cache dir: %s", err))
	}

	tempDir = filepath.Join(RuckstackWorkDir, "tmp", "ruckstack-run-"+strconv.FormatInt(int64(rand.Int()), 10))

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		panic(fmt.Sprintf("Cannot create temp dir: %s", err))
	}
	ui.VPrintf("Ruckstack temp dir: %s", tempDir)
}

func Cleanup() {
	err := os.RemoveAll(tempDir)
	ui.VPrintf("Error deleting %s: %s", tempDir, err)
}

/**
Returns the given path as a sub-path of the Ruckstack "temporary" directory.
Any "*" in the path will be replaced with a random value
*/
func TempPath(pathInTmp string) string {
	pathInTmp = strings.Replace(pathInTmp, "*", strconv.Itoa(rand.Int()), 1)
	return filepath.Join(tempDir, pathInTmp)
}

/**
Returns the given path as a sub-path of the Ruckstack "cache" dir.
The cache directory is preserved from one run to the next
*/
func CachePath(pathInCache string) string {
	return filepath.Join(cacheRoot, pathInCache)
}

/**
Returns the given path as a sub-path of the Ruckstack "out" dir.
*/
func OutPath(path string) string {
	if OutDir == "" {
		ui.Fatal("out directory not specified")
	}
	return filepath.Join(OutDir, path)
}
