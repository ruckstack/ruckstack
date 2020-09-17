package helm

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/kubeclient"
	"os"
	"os/exec"
	"path/filepath"
)

func ExecHelm(args ...string) error {
	command := exec.Command(common2.InstallDir()+"/lib/helm", args...)
	command.Env = append(command.Env,
		fmt.Sprintf("KUBECONFIG=%s", kubeclient.KubeconfigFile()),
		fmt.Sprintf("HELM_HOME=%s", filepath.Join(common2.InstallDir(), "data", "helm_home")),
	)
	command.Dir = common2.InstallDir()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
