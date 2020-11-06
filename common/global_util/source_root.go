package global_util

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"path/filepath"
	"strings"
)

var sourceRoot string
var isRunningTests *bool

func IsRunningTests() bool {
	if isRunningTests == nil {
		for _, val := range os.Args {
			if val == "--server-home" {
				newValue := false
				isRunningTests = &newValue

				return *isRunningTests
			}
		}

		var newValue bool
		executable, err := os.Executable()
		if err != nil {
			ui.Fatalf("Cannot determine if we are running tests: %s", err)
		} else {
			newValue = strings.HasPrefix(executable, "/tmp/__") || strings.HasPrefix(executable, "/tmp/go-build")
		}

		isRunningTests = &newValue
	}

	return *isRunningTests
}

/**
Returns the source root. Panic's if used outside a test
*/
func GetSourceRoot() string {
	if sourceRoot == "" {
		if !IsRunningTests() {
			panic("Unexpected request for source root")
		}

		var err error
		//work from the working directory
		sourceRoot, err = os.Getwd()

		if err != nil {
			ui.Fatalf("Cannot determine working directory: %s", err)
		}

		//search back until we fine the root containing the LICENSE file
		for sourceRoot != "/" {
			if _, err := os.Stat(filepath.Join(sourceRoot, "LICENSE")); os.IsNotExist(err) {
				sourceRoot = filepath.Dir(sourceRoot)
				continue
			}
			break
		}

		if sourceRoot == "/" {
			ui.Fatalf("Cannot determine source root")
		}
		ui.VPrintf("Source root: %s\n", sourceRoot)
	}

	return sourceRoot
}
