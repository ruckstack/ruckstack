package k3s

import (
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"os"
	"os/exec"
)

func ExecCrictl(args ...string) error {
	command := exec.Command(common2.InstallDir()+"/lib/k3s", append([]string{"crictl"}, args...)...)
	command.Dir = common2.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
