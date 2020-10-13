package agent

import (
	"context"
	"github.com/rancher/k3s/pkg/agent"
	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/containerd"
)

func Start() error {
	agentConfig := cmds.Agent{
		Token:                    "",
		TokenFile:                "",
		ClusterSecret:            "",
		ServerURL:                "https://localhost:6443",
		DisableLoadBalancer:      false,
		ResolvConf:               "",
		DataDir:                  environment.ServerHome + "/data",
		NodeIP:                   "",
		NodeExternalIP:           environment.LocalConfig.BindAddress,
		NodeName:                 "",
		PauseImage:               "",
		Snapshotter:              "",
		Docker:                   false,
		ContainerRuntimeEndpoint: containerd.SocketFile,
		NoFlannel:                false,
		FlannelIface:             environment.LocalConfig.BindAddressInterface,
		FlannelConf:              environment.ServerHome + "/config/flannel.env",
		Debug:                    false,
		Rootless:                 false,
		RootlessAlreadyUnshared:  false,
		WithNodeID:               false,
		EnableSELinux:            false,
		ProtectKernelDefaults:    false,
		AgentShared:              cmds.AgentShared{},
		ExtraKubeletArgs:         []string{"root-dir=" + environment.ServerHome + "/data/kubelet"},
		ExtraKubeProxyArgs:       nil,
		Labels:                   nil,
		Taints:                   nil,
		PrivateRegistry:          "",
	}

	return agent.Run(context.Background(), agentConfig)
}
