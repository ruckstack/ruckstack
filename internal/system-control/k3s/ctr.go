package k3s

import (
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"os"
	"os/exec"
)

func ExecCtr(args ...string) {
	command := exec.Command(util.InstallDir()+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = util.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		panic(err)
	}
}
