package environment

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"path/filepath"
	"strings"
)

var (
	ruckstackHome string
	resourceRoot  string
	cacheRoot     string
)

func init() {
	isRunningTests := false
	executable, err := os.Executable()
	if err == nil {
		ui.VPrintln("Cannot determine if we are running tests: ", err)
	}
	isRunningTests = strings.HasPrefix(executable, "/tmp/")

	//find ruckstackHome
	if isRunningTests {
		//work from the working directory
		ruckstackHome, err = os.Getwd()

		if err != nil {
			ui.Fatalf("Cannot determine working directory: %s", err)
		}
	} else {
		ruckstackHome = filepath.Dir(executable)
	}

	//search back until we fine the root containing the LICENSE file
	for ruckstackHome != "/" {
		if _, err := os.Stat(filepath.Join(ruckstackHome, "LICENSE")); os.IsNotExist(err) {
			ruckstackHome = filepath.Dir(ruckstackHome)
			continue
		}
		break
	}

	if ruckstackHome == "/" {
		ui.Fatal("Cannot determine Ruckstack home")
	}
	ui.VPrintf("Ruckstack home: %s\n", ruckstackHome)

	//find resourceRoot
	if isRunningTests {
		resourceRoot = ruckstackHome + "/builder/cli/install_root/resources"
	} else {
		resourceRoot = ruckstackHome + "/resources"
	}
	ui.VPrintf("Ruckstack resource root: %s", resourceRoot)

	//find cacheRoot
	cacheRoot = os.Getenv("RUCKSTACK_CACHE_DIR")
	if cacheRoot == "" {
		cacheRoot = ruckstackHome + "/cache"
	}
	ui.VPrintf("Ruckstack cache root: %s", cacheRoot)
}

/**
Returns the full path to the given subpath of "resources" in ruckstackHome.
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
*/
func TempPath(pathInTmp string) string {
	return filepath.Join(resourceRoot, "tmp", pathInTmp)
}

/**
Returns the given path as a sub-path of the Ruckstack "cache" dir.
The cache directory is preserved from one run to the next
*/
func CachePath(pathInCache string) string {
	return filepath.Join(cacheRoot, pathInCache)
}
