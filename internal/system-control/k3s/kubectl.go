package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"os"
	"os/exec"
)

func ExecKubectl(args ...string) {
	command := exec.Command(util.InstallDir()+"/lib/k3s", append([]string{"kubectl"}, args...)...)
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
	)
	command.Dir = util.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	_ = command.Run()

}
