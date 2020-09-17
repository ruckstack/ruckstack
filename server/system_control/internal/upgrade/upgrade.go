package upgrade

import (
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"os"
	"os/exec"
)

func Upgrade(upgradeFile string) error {

	command := exec.Command(upgradeFile, "--upgrade", common2.InstallDir())
	command.Dir = common2.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
