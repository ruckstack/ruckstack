package k3s

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"os/exec"
)

func ExecCrictl(serverHome string, args ...string) error {
	command := exec.Command(serverHome+"/lib/k3s", append([]string{"crictl"}, args...)...)
	command.Dir = serverHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	return command.Run()
}
