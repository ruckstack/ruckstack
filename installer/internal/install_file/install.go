package install_file

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type InstallOptions struct {
	AdminGroup  string
	BindAddress string
	JoinToken   string
	TargetDir   string
}

/**
Installs this file
*/
func (installFile *InstallFile) Install(installOptions InstallOptions) error {

	ui.Printf("%s %s Installer\n", installFile.PackageConfig.Name, installFile.PackageConfig.Version)
	ui.Println("---------------------------------------------")

	clusterConfig := new(config.ClusterConfig)
	localConfig := new(config.LocalConfig)

	if installOptions.TargetDir == "" {
		installOptions.TargetDir = askInstallPath()
	}

	var err error

	_, err = config.LoadPackageConfig(installOptions.TargetDir)
	if err == nil {
		ui.VPrintf("%s is an existing install. Upgrading...", installOptions.TargetDir)
		return installFile.Upgrade(installOptions)
	} else if !os.IsNotExist(err) {
		ui.Fatalf("Error checking path %s: %s", installOptions.TargetDir, err)
	}

	shouldJoinCluster := false
	if installOptions.JoinToken == "none" {
		shouldJoinCluster = false
	} else if installOptions.JoinToken == "" {
		shouldJoinCluster = ui.PromptForBoolean("Join an existing cluster", nil)
	} else {
		shouldJoinCluster = true
	}

	var addNodeToken *config.AddNodeToken
	if shouldJoinCluster {
		addNodeToken, err = joinCluster(installOptions.JoinToken, installFile)
		if err != nil {
			return fmt.Errorf("error joining cluster: %s", err)
		}
	}

	if installOptions.BindAddress == "" {
		installOptions.BindAddress, err = askBindAddress()
		if err != nil {
			return err
		}
	}

	if installOptions.AdminGroup == "" {
		installOptions.AdminGroup = askAdminGroup()
	}

	localConfig.AdminGroup = installOptions.AdminGroup
	localConfig.BindAddress = installOptions.BindAddress
	if addNodeToken != nil {
		localConfig.Join.Server = addNodeToken.Server
		localConfig.Join.Token = addNodeToken.Token
	}

	defer ui.StartProgressf("Installing to %s", installOptions.TargetDir).Stop()

	if err := installFile.Extract(installOptions.TargetDir, localConfig); err != nil {
		return err
	}

	if err := localConfig.Save(installOptions.TargetDir, installFile.PackageConfig); err != nil {
		return err
	}

	if err := clusterConfig.Save(installOptions.TargetDir, installFile.PackageConfig, localConfig); err != nil {
		return err
	}

	ui.Println("\n\nInstallation complete")
	ui.Printf("To start the server, run `%s/bin/%s start`\n\n", installOptions.TargetDir, installFile.SystemConfig.ManagerFilename)

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

func askInstallPath() string {
	longPathCheck := func(input string) error {
		absInstallPath, _ := filepath.Abs(input)

		if len(absInstallPath) > 50 {
			//if install path is too long, socket paths get longer than the 107 chars linux supports
			return fmt.Errorf("%s is too deeply nested. Choose a different directory", absInstallPath)
		}

		return nil
	}

	return ui.PromptForString("Install path", "", longPathCheck, ui.NotDirectoryCheck)
}

func joinCluster(joinToken string, installFile *InstallFile) (*config.AddNodeToken, error) {
	addNodeToken := &config.AddNodeToken{}

	var parseTokenCheck = func(token string) error {
		token = strings.TrimSpace(token)
		joinTokenYaml, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return fmt.Errorf("error parsing token: %s", err)
		}

		if len(joinTokenYaml) == 0 {
			return fmt.Errorf("could not read data in token")
		}

		tokenDecoder := yaml.NewDecoder(bytes.NewReader(joinTokenYaml))
		if err = tokenDecoder.Decode(addNodeToken); err != nil {
			return fmt.Errorf("error decoding token: %s", err)
		}

		return nil
	}

	if joinToken == "" {
		joinToken = ui.PromptForString(fmt.Sprintf("Run `%s cluster add-node` on the primary machine in the cluster and enter the token here", installFile.SystemConfig.ManagerFilename), "", parseTokenCheck)
	}

	timeout := time.Second
	for _, serverPortAndProtocol := range []string{"tcp/6443", "udp/8472"} {
		splitInfo := strings.Split(serverPortAndProtocol, "/")
		conn, err := net.DialTimeout(splitInfo[0], net.JoinHostPort(addNodeToken.Server, splitInfo[1]), timeout)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to server: %s", err)
		}
		conn.Close()
	}

	return addNodeToken, nil

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
