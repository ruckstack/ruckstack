package server

import (
	"fmt"
	"github.com/k3d-io/k3d/v5/cmd/cluster"
	"github.com/k3d-io/k3d/v5/cmd/registry"
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
	"net"
	"os/user"
	"path/filepath"
)

type InstallOptions struct {
	AdminGroup string
	//BindAddress string
	JoinToken string
}

func Setup(installOptions InstallOptions) error {
	if config.LocalConfig != nil {
		ui.Fatalf("Server already configured. Start with `system-control start`")
	}

	var err error

	ui.Printf("%s %s Setup\n", config.PackageConfig.Name, config.PackageConfig.Version)
	ui.Println("---------------------------------------------")

	clusterConfig := new(config.ClusterConfigType)
	localConfig := new(config.LocalConfigType)

	shouldJoinCluster := false
	if installOptions.JoinToken == "none" {
		shouldJoinCluster = false
	} else if installOptions.JoinToken == "" {
		shouldJoinCluster = ui.PromptForBoolean("Join an existing cluster", nil)
	} else {
		shouldJoinCluster = true
	}

	//var addNodeToken *packageConfig.AddNodeToken
	if shouldJoinCluster {
		//	addNodeToken, err = joinCluster(installOptions.JoinToken, installFile)
		//	if err != nil {
		//		return fmt.Errorf("error joining cluster: %s", err)
		//	}
	}

	//if installOptions.BindAddress == "" {
	//	installOptions.BindAddress, err = askBindAddress()
	//	if err != nil {
	//		return err
	//	}
	//}

	//if installOptions.AdminGroup == "" {
	//	installOptions.AdminGroup = askAdminGroup()
	//}

	localConfig.AdminGroup = installOptions.AdminGroup
	//localConfig.BindAddress = installOptions.BindAddress
	//if addNodeToken != nil {
	//	localConfig.Join.Server = addNodeToken.Server
	//	localConfig.Join.Token = addNodeToken.Token
	//}

	createRegistryCmd := registry.NewCmdRegistryCreate()
	createRegistryCmd.SetArgs([]string{
		config.PackageConfig.Id + "-registry.localhost",
		"--port", "5000",
		"-v", filepath.FromSlash(config.ServerHome+"/registry") + ":/var/lib/registry",
	})
	err = createRegistryCmd.Execute()

	createCmd := cluster.NewCmdClusterCreate()
	createCmd.SetArgs([]string{
		config.PackageConfig.Id,
		"--k3s-arg", "--cluster-init@server:0",
		"--servers", "1",
		"--registry-use", "k3d-" + config.PackageConfig.Id + "-registry.localhost:5000",
		"-v", filepath.FromSlash(config.ServerHome+"/data") + ":/data",
		"--wait",
	})
	err = createCmd.Execute()

	if err != nil {
		ui.Fatalf("error!", err)
	}

	if err := localConfig.Save(); err != nil {
		return err
	}

	if err := clusterConfig.Save(); err != nil {
		return err
	}

	ui.Printf("\n\n%s setup complete and started\n", config.PackageConfig.Name)

	return nil
}

func askAdminGroup() string {

	var validGroup = func(input string) error {
		foundGroup, err := user.LookupGroup(input)
		if err == nil {
			if foundGroup.Gid == "0" {
				return fmt.Errorf("must specify a non-root group")
			} else {
				return nil
			}
		} else {
			return fmt.Errorf("Unknown group: %s", input)
		}

	}

	return ui.PromptForString("Administrator group", "", validGroup)

}

func askBindAddress() (string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("cannot determine network interfaces: %s", err)
	}

	var validInterfaceCheck = func(input string) error {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				return err
			}

			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil && ip.String() == input {
					return nil
				}
			}
		}

		return fmt.Errorf("Invalid IP address for this machine: %s", input)
	}

	return ui.PromptForString("IP address to bind to", "", validInterfaceCheck), nil
}
