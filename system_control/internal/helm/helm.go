package helm

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common"
	"github.com/ruckstack/ruckstack/common/kubeclient"
	"os"
	"os/exec"
	"path/filepath"
)

func ExecHelm(args ...string) error {
	command := exec.Command(common.InstallDir()+"/lib/helm", args...)
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
		fmt.Sprintf("HELM_HOME=%s", filepath.Join(common.InstallDir(), "data", "helm_home")),
	)
	command.Dir = common.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
