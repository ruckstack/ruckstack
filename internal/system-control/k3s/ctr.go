package k3s

import (
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"io"
	"log"
	"os/exec"
)

func ExecCtr(stdout io.Writer, stderr io.Writer, args ...string) {

	command := exec.Command(util.InstallDir()+"/lib/k3s", append([]string{"ctr"}, args...)...)
	command.Dir = util.InstallDir()
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		log.Printf("Cannot import images %s: %s", args, err)
	}
}
