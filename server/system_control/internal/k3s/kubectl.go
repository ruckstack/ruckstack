package k3s

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/kube"
	"os"
	"os/exec"
)

func ExecKubectl(serverHome string, args ...string) error {
	command := exec.Command(serverHome+"/lib/k3s", append([]string{"kubectl"}, args...)...)
	command.Env = os.Environ()
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kube.KubeconfigFile),
	)
	command.Dir = serverHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	return command.Run()

}
