package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"io"
	"os/exec"
)

func ExecCtr(stdout io.Writer, stderr io.Writer, args ...string) error {

	command := exec.Command(util.InstallDir()+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = util.InstallDir()
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("Cannot import images %s: %s", args, err)
	}
	return nil
}
