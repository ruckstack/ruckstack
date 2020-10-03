package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"os/exec"
)

func ExecCtr(serverHome string, args ...string) error {

	command := exec.Command(serverHome+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = serverHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	if err := command.Run(); err != nil {
		return fmt.Errorf("Cannot import images %s: %s", args, err)
	}
	return nil
}
