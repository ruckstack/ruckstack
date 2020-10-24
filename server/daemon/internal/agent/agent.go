package agent

import (
	"context"
	"github.com/rancher/k3s/pkg/agent"
	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/daemon/internal/containerd"
	"github.com/ruckstack/ruckstack/server/daemon/internal/k3s"
	"github.com/ruckstack/ruckstack/server/internal/environment"
)

func Start(parent context.Context) error {
	agentConfig := cmds.Agent{
		ServerURL:                "https://127.0.0.1:6443",
		DataDir:                  environment.ServerHome + "/data",
		NodeExternalIP:           environment.LocalConfig.BindAddress,
		ContainerRuntimeEndpoint: containerd.SocketFile,
		FlannelIface:             environment.LocalConfig.BindAddressInterface,
		FlannelConf:              environment.ServerHome + "/config/flannel.env",
		ExtraKubeletArgs:         []string{"root-dir=" + environment.ServerHome + "/data/kubelet"},
		Debug:                    true,
		Token:                    k3s.ServerToken,
		DisableLoadBalancer:      true,

		TokenFile:               "",
		ClusterSecret:           "",
		ResolvConf:              "",
		NodeIP:                  "",
		NodeName:                "",
		PauseImage:              "",
		Snapshotter:             "",
		Docker:                  false,
		NoFlannel:               false,
		Rootless:                false,
		RootlessAlreadyUnshared: false,
		WithNodeID:              false,
		EnableSELinux:           false,
		ProtectKernelDefaults:   false,
		AgentShared:             cmds.AgentShared{},
		ExtraKubeProxyArgs:      nil,
		Labels:                  nil,
		Taints:                  nil,
		PrivateRegistry:         "",
	}

	cmds.LogConfig.LogFile = environment.ServerHome + "/logs/k3s.agent.log"
	cmds.LogConfig.AlsoLogToStderr = false

	ui.Println("Starting agent...")
	defer ui.Println("Starting agent...DONE")
	go func() {
		err := agent.Run(parent, agentConfig)
		ui.Fatalf("agent exited: %s, ", err)
	}()

	return nil
}
