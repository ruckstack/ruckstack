package k3s

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"io"
	"os/exec"
)

func ExecCtr(stdout io.Writer, stderr io.Writer, args ...string) error {

	command := exec.Command(common2.InstallDir()+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = common2.InstallDir()
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("Cannot import images %s: %s", args, err)
	}
	return nil
}
