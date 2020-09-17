package environment

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
	"path/filepath"
)

var (
	ruckstackHome string
)

func GetRuckstackHome() string {
	if ruckstackHome != "" {
		return ruckstackHome
	}

	defaultHome := "/ruckstack"

	ruckstackHome = defaultHome
	_, err := os.Stat(ruckstackHome)
	if err == nil {
		ui.VPrintf("Ruckstack home: %s\n", ruckstackHome)
		return ruckstackHome
	}

	//No /ruckstack directory. Figure out the home directory
	executable, err := os.Executable()
	if err != nil {
		ui.Printf("Cannot determine executable. Using default home directory. Error: %s\n", err)
		return defaultHome
	}
	if executable == "ruckstack" {
		ruckstackHome = filepath.Dir(executable)
	} else {
		ruckstackHome, err = os.Getwd()

		if err != nil {
			ui.Printf("Cannot determine working directory. Using default home directory. Error: %s\n", err)
			return defaultHome
		}
	}

	for ruckstackHome != "/" {
		if _, err := os.Stat(filepath.Join(ruckstackHome, "LICENSE")); os.IsNotExist(err) {
			ruckstackHome = filepath.Dir(ruckstackHome)
			continue
		}
		break
	}

	if ruckstackHome == "/" {
		ui.VPrintf("Cannot determine Ruckstack home. Using default")
		ruckstackHome = defaultHome
	}

	ui.VPrintf("Ruckstack home: %s\n", ruckstackHome)
	return ruckstackHome

}

/**
Returns the given path as a sub-path of the Ruckstack "tmp" dir
*/
func TempPath(pathInTmp string) string {
	return filepath.Join(GetRuckstackHome(), "tmp", pathInTmp)
}
