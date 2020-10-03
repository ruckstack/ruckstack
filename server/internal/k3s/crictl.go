package k3s

import (
	"os"
	"os/exec"
)

func ExecCrictl(serverHome string, args ...string) error {
	command := exec.Command(serverHome+"/lib/k3s", append([]string{"crictl"}, args...)...)
	command.Dir = serverHome
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
