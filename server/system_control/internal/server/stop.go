package server

import (
	"github.com/k3d-io/k3d/v5/cmd/node"
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
)

func Stop() error {
	for _, nodeName := range []string{
		"k3d-" + config.PackageConfig.Id + "-server-0",
		"k3d-" + config.PackageConfig.Id + "-serverlb",
		"k3d-" + config.PackageConfig.Id + "-tools",
		"k3d-" + config.PackageConfig.Id + "-registry.localhost",
	} {
		createCmd := node.NewCmdNodeStop()
		createCmd.SetArgs([]string{nodeName})

		err := createCmd.Execute()
		if err != nil {
			return err
		}
	}

	return nil

}
