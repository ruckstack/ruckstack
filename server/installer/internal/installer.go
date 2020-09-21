package internal

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/files"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func Install(packageConfig *config.PackageConfig, installerArgs *InstallerArgs, zipReader *zip.ReadCloser) error {

	fmt.Printf("%s %s Installer\n", packageConfig.Name, packageConfig.Version)
	fmt.Println("---------------------------------------------")

	systemConfig := new(config.SystemConfig)
	localConfig := new(config.LocalConfig)

	ui := bufio.NewScanner(os.Stdin)

	installPath := getInstallPath(ui, packageConfig, installerArgs)
	joinToken, err := getJoinToken(ui, packageConfig, installerArgs)
	if err != nil {
		return err
	}

	bindAddress, err := getBindAddress(ui, installerArgs)
	if err != nil {
		return err
	}

	adminGroup := getAdminGroup(ui, installerArgs)

	localConfig.AdminGroup = adminGroup.Name
	localConfig.BindAddress = bindAddress
	if joinToken != nil {
		localConfig.Join.Server = joinToken.Server
		localConfig.Join.Token = joinToken.Token
	}

	environment.SetPackageConfig(packageConfig)
	environment.SetSystemConfig(systemConfig)
	environment.SetLocalConfig(localConfig)

	fmt.Printf("Installing to %s...\n", installPath)

	if err := global_util.Unzip(installPath, zipReader); err != nil {
		return err
	}

	//TODO: Re-enable permission check
	//if err := files.CheckFilePermissions(installPath, file.Name); err != nil {
	//	return err
	//}

	if err := os.MkdirAll(installPath+"/config", 0755); err != nil {
		return err
	}

	if joinToken != nil {
		if err := ioutil.WriteFile(filepath.Join(installPath, "config/kubeconfig.yaml"), []byte(joinToken.KubeConfig), 0640); err != nil {
			return nil
		}
	}

	systemConfigFile, err := os.OpenFile(installPath+"/config/system.config", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	systemConfigEncoder := yaml.NewEncoder(systemConfigFile)
	if err := systemConfigEncoder.Encode(systemConfig); err != nil {
		return err
	}
	if err := files.CheckFilePermissions(installPath, "config/system.config"); err != nil {
		return err
	}

	localConfigFile, err := os.OpenFile(installPath+"/config/local.config", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	localConfigEncoder := yaml.NewEncoder(localConfigFile)
	if err := localConfigEncoder.Encode(localConfig); err != nil {
		return err
	}
	if err := files.CheckFilePermissions(installPath, "config/local.config"); err != nil {
		return err
	}

	fmt.Println("\n\nInstallation complete")
	fmt.Printf("To start the server, run `%s/bin/%s start`\n\n", installPath, packageConfig.SystemControlName)

	return nil
}

func getAdminGroup(ui *bufio.Scanner, installerArgs *InstallerArgs) *user.Group {
	var enteredGroup string

	if installerArgs != nil {
		enteredGroup = installerArgs.AdminGroup
	}

	if enteredGroup == "" {
		fmt.Print("Administrator group: ")
		ui.Scan()
		enteredGroup = ui.Text()
	}

	foundGroup, err := user.LookupGroup(enteredGroup)
	if err == nil {
		if foundGroup.Gid == "0" {
			fmt.Println("Must specify a non-root group")
			return getAdminGroup(ui, nil)
		} else {
			return foundGroup
		}
	} else {
		fmt.Println("Invalid group: " + enteredGroup)
		return getAdminGroup(ui, nil)
	}
}

func getInstallPath(ui *bufio.Scanner, packageConfig *config.PackageConfig, installerArgs *InstallerArgs) string {
	var installPath string
	if installerArgs != nil {
		installPath = installerArgs.InstallPath
	}

	if installPath == "" {
		fmt.Print("Install path: ")
		ui.Scan()
		installPath = ui.Text()
	}

	absInstallPath, err := filepath.Abs(installPath)

	if len(absInstallPath) > 50 {
		//if install path is too long, socket paths get longer than the 107 chars linux supports
		fmt.Println(absInstallPath + " is too deeply nested. Choose a different directory.")
		return getInstallPath(ui, nil, nil)
	}

	stat, err := os.Stat(absInstallPath)
	if os.IsNotExist(err) {
		//that is what we want
	} else {
		if !stat.IsDir() {
			fmt.Println(absInstallPath + " is not a directory")
			return getInstallPath(ui, nil, nil)
		}
	}

	stat, err = os.Stat(filepath.Join(absInstallPath, ".package.config"))
	if os.IsNotExist(err) {
		//that is what we want
	} else {
		fmt.Printf(absInstallPath+" already exists\n\nTo upgrade, run `%s/bin/%s upgrade --file %s`\n", absInstallPath, packageConfig.SystemControlName, installPath)
		os.Exit(1)
	}

	return absInstallPath
}

func getJoinToken(ui *bufio.Scanner, packageConfig *config.PackageConfig, installerArgs *InstallerArgs) (*config.AddNodeToken, error) {
	addNodeToken := &config.AddNodeToken{}

	var joinToken string
	if installerArgs != nil {
		joinToken = installerArgs.JoinToken
	}

	var joinCluster bool
	if joinToken == "" {
		gotValidResponse := false

		for !gotValidResponse {
			fmt.Print("Join an existing cluster? [y|n] ")
			ui.Scan()
			joinResponse := strings.ToLower(ui.Text())

			if joinResponse == "y" {
				gotValidResponse = true
				joinCluster = true
			} else if joinResponse == "n" {
				gotValidResponse = true
				return nil, nil
			}
		}
	}

	for joinCluster && joinToken == "" {
		fmt.Printf("Run `%s cluster add-node` on the primary machine in the cluster and enter the token here:\n", packageConfig.SystemControlName)

		ui.Scan()
		joinToken = strings.TrimSpace(ui.Text())

		joinTokenYaml, err := base64.StdEncoding.DecodeString(joinToken)
		if err != nil {
			fmt.Printf("Error parsing token\n\n")
			joinToken = ""
			continue
		}

		tokenDecoder := yaml.NewDecoder(bytes.NewReader(joinTokenYaml))
		err = tokenDecoder.Decode(addNodeToken)

		if err != nil {
			fmt.Printf("Error reading token\n\n")
			joinToken = ""
			continue
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
	}

	return addNodeToken, nil

}

func getBindAddress(ui *bufio.Scanner, installerArgs *InstallerArgs) (string, error) {
	var bindAddress string
	if installerArgs != nil {
		bindAddress = installerArgs.BindAddress
	}

	if bindAddress == "" {
		fmt.Print("IP address to bind to: ")
		ui.Scan()
		bindAddress = ui.Text()
	}

	foundIp := false
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.String() == bindAddress {
				foundIp = true
				break
			}
		}
	}

	if !foundIp {
		fmt.Println("Unknown IP address " + bindAddress)
		return getBindAddress(ui, nil)
	}

	return bindAddress, nil
}

type InstallerArgs struct {
	InstallPath string
	AdminGroup  string
	BindAddress string
	JoinServer  string
	JoinToken   string
}
