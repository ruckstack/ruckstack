package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common"
	"io"
	"os/exec"
)

func ExecCtr(stdout io.Writer, stderr io.Writer, args ...string) error {

	command := exec.Command(common.InstallDir()+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = common.InstallDir()
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("Cannot import images %s: %s", args, err)
	}
	return nil
}
