package k3s

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"os/exec"
)

func ExecCrictl(serverHome string, args ...string) error {
	command := exec.Command(serverHome+"/lib/k3s", append([]string{"crictl", "--data-dir", environment.ServerHome + "/data"}, args...)...)
	command.Dir = serverHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	return command.Run()
}
