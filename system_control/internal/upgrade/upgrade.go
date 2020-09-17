package upgrade

import (
	"github.com/ruckstack/ruckstack/common"
	"os"
	"os/exec"
)

func Upgrade(upgradeFile string) error {

	command := exec.Command(upgradeFile, "--upgrade", common.InstallDir())
	command.Dir = common.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
