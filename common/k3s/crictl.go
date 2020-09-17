package k3s

import (
	"github.com/ruckstack/ruckstack/common"
	"os"
	"os/exec"
)

func ExecCrictl(args ...string) error {
	command := exec.Command(common.InstallDir()+"/lib/k3s", append([]string{"crictl"}, args...)...)
	command.Dir = common.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
