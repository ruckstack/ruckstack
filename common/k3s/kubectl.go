package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common"
	"github.com/ruckstack/ruckstack/common/kubeclient"
	"os"
	"os/exec"
)

func ExecKubectl(args ...string) error {
	command := exec.Command(common.InstallDir()+"/lib/k3s", append([]string{"kubectl"}, args...)...)
	command.Env = os.Environ()
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
	)
	command.Dir = common.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
