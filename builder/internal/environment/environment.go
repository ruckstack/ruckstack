package environment

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	RuckstackHome string
	resourceRoot  string
	cacheRoot     string
	tempDir       string

	OutDir     string
	ProjectDir string

	isRunningTests = false
)

func init() {
	executable, err := os.Executable()
	if err != nil {
		ui.VPrintln("Cannot determine if we are running tests: ", err)
	} else {
		isRunningTests = strings.HasPrefix(executable, "/tmp/")
	}

	//find RuckstackHome
	if isRunningTests {
		//work from the working directory
		RuckstackHome, err = os.Getwd()

		if err != nil {
			ui.Fatalf("Cannot determine working directory: %s", err)
		}
	} else {
		RuckstackHome = filepath.Dir(executable)
	}

	//search back until we fine the root containing the LICENSE file
	for RuckstackHome != "/" {
		if _, err := os.Stat(filepath.Join(RuckstackHome, "LICENSE")); os.IsNotExist(err) {
			RuckstackHome = filepath.Dir(RuckstackHome)
			continue
		}
		break
	}

	if RuckstackHome == "/" {
		panic("Cannot determine Ruckstack home")
	}
	ui.VPrintf("Ruckstack home: %s\n", RuckstackHome)

	//find resourceRoot
	if isRunningTests {
		resourceRoot = RuckstackHome + "/builder/cli/install_root/resources"
	} else {
		resourceRoot = RuckstackHome + "/resources"
	}
	ui.VPrintf("Ruckstack resource root: %s", resourceRoot)

	//find cacheRoot
	if isRunningTests {
		cacheRoot = RuckstackHome + "/cache"
	} else {
		cacheRoot = os.Getenv("RUCKSTACK_CACHE_DIR")
		if cacheRoot == "" {
			cacheRoot = "/data/cache"
		}
	}
	ui.VPrintf("Ruckstack cache root: %s", cacheRoot)

	tempDir = os.Getenv("RUCKSTACK_TEMP_DIR")
	if tempDir == "" {
		//find tempDir
		if isRunningTests {
			tempDir = RuckstackHome + "/tmp"
		} else {
			tempDir = "/data/tmp/"
		}
	}
	tempDir = filepath.Join(tempDir, "ruckstack-run-"+strconv.FormatInt(int64(rand.Int()), 10))

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ui.Fatalf("Cannot create temp dir: %s", err)
	}
	ui.VPrintf("Ruckstack temp dir: %s", tempDir)
}

/**
Returns true if ruckstack is running via the launcher
*/
func IsRunningLauncher() bool {
	return os.Getenv("RUCKSTACK_DOCKERIZED") == "true"
}

/**
Returns the full path to the given subpath of "resources" in RuckstackHome.
Returns an error if the file does not exist
*/
func ResourcePath(path string) (string, error) {
	resourcePath := filepath.Join(resourceRoot, path)

	if _, err := os.Stat(resourcePath); err != nil {
		return "", err
	}

	return resourcePath, nil
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
