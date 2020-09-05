package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"os"
	"os/exec"
)

func ExecKubectl(args ...string) error {
	command := exec.Command(util.InstallDir()+"/lib/k3s", append([]string{"kubectl"}, args...)...)
	command.Env = os.Environ()
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
	)
	command.Dir = util.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
