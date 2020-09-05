package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/kubeclient"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"os"
	"os/exec"
	"path/filepath"
)

func ExecHelm(args ...string) error {
	command := exec.Command(util.InstallDir()+"/lib/helm", args...)
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
		fmt.Sprintf("HELM_HOME=%s", filepath.Join(util.InstallDir(), "data", "helm_home")),
	)
	command.Dir = util.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
