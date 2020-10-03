package upgrade

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"os"
	"os/exec"
)

func Upgrade(upgradeFile string) error {

	command := exec.Command(upgradeFile, "--upgrade", environment.ServerHome)
	command.Dir = environment.ServerHome
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
