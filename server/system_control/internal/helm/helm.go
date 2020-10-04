package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/kubeclient"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"os/exec"
	"path/filepath"
)

func ExecHelm(args ...string) error {
	command := exec.Command(environment.ServerHome+"/lib/helm", args...)
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile(environment.ServerHome)),
		fmt.Sprintf("HELM_HOME=%s", filepath.Join(environment.ServerHome, "data", "helm_home")),
	)
	command.Dir = environment.ServerHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()

	return command.Run()
}
