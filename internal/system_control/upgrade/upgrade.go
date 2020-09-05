package upgrade

import (
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"os"
	"os/exec"
)

func Upgrade(upgradeFile string) {

	command := exec.Command(upgradeFile, "--upgrade", util.InstallDir())
	command.Dir = util.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		panic(err)
	}

}
