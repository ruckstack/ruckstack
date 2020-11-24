package upgrade

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"os/exec"
)

func Upgrade(upgradeFile string) error {

	command := exec.Command(upgradeFile, "--upgrade", environment.ServerHome)
	command.Dir = environment.ServerHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	return command.Run()

}
