package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"os"
	"os/exec"
)

func ExecKubectl(args ...string) error {
	command := exec.Command(environment.ServerHome+"/lib/k3s", append([]string{"--data-dir", environment.ServerHome + "/data/kubectl", "kubectl"}, args...)...)
	command.Env = os.Environ()
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kube.KubeconfigFile),
	)
	command.Dir = environment.ServerHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	return command.Run()

}
