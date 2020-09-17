package k3s

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/kubeclient"
	"os"
	"os/exec"
)

func ExecKubectl(args ...string) error {
	command := exec.Command(common2.InstallDir()+"/lib/k3s", append([]string{"kubectl"}, args...)...)
	command.Env = os.Environ()
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
	)
	command.Dir = common2.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()

}
