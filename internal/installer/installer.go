package installer

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/system-control/files"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"gopkg.in/yaml.v2"
)

func Install(packageConfig *internal.PackageConfig, installerArgs *InstallerArgs, zipReader *zip.ReadCloser) {

	fmt.Printf("%s %s Installer\n", packageConfig.Name, packageConfig.Version)
	fmt.Println("---------------------------------------------")

	systemConfig := new(internal.SystemConfig)
	localConfig := new(internal.LocalConfig)

	ui := bufio.NewScanner(os.Stdin)

	installPath := getInstallPath(ui, packageConfig, installerArgs)
	joinToken := getJoinToken(ui, packageConfig, installerArgs)
	bindAddress := getBindAddress(ui, installerArgs)
	adminGroup := getAdminGroup(ui, installerArgs)

	localConfig.AdminGroup = adminGroup.Name
	localConfig.BindAddress = bindAddress
	if joinToken != nil {
		localConfig.Join.Server = joinToken.Server
		localConfig.Join.Token = joinToken.Token
	}

	util.SetPackageConfig(packageConfig)
	util.SetSystemConfig(systemConfig)
	util.SetLocalConfig(localConfig)

	fmt.Printf("Installing to %s...\n", installPath)

	err := extract(installPath, zipReader)
	if err != nil {
		panic(err)
	}

	util.Check(os.MkdirAll(installPath+"/config", 0755))

	if joinToken != nil {
		util.Check(ioutil.WriteFile(filepath.Join(installPath, "config/kubeconfig.yaml"), []byte(joinToken.KubeConfig), 0640))
	}

	systemConfigFile, err := os.OpenFile(installPath+"/config/system.config", os.O_CREATE|os.O_RDWR, 0644)
	systemConfigEncoder := yaml.NewEncoder(systemConfigFile)
	util.Check(systemConfigEncoder.Encode(systemConfig))
	util.Check(files.CheckFilePermissions(installPath, "config/system.config"))

	localConfigFile, err := os.OpenFile(installPath+"/config/local.config", os.O_CREATE|os.O_RDWR, 0644)
	localConfigEncoder := yaml.NewEncoder(localConfigFile)
	util.Check(localConfigEncoder.Encode(localConfig))
	util.Check(files.CheckFilePermissions(installPath, "config/local.config"))

	fmt.Println("\n\nInstallation complete")
	fmt.Printf("To start the server, run `%s/bin/%s start`\n\n", installPath, packageConfig.SystemControlName)
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

func getInstallPath(ui *bufio.Scanner, packageConfig *internal.PackageConfig, installerArgs *InstallerArgs) string {
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

func getJoinToken(ui *bufio.Scanner, packageConfig *internal.PackageConfig, installerArgs *InstallerArgs) *AddNodeToken {
	addNodeToken := &AddNodeToken{}

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
				return nil
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
				fmt.Printf("Cannot connect to server -- %s\n", err.Error())

				util.Check(err)
			}
			conn.Close()
		}
	}

	return addNodeToken

}

func getBindAddress(ui *bufio.Scanner, installerArgs *InstallerArgs) string {
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
	util.Check(err)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		util.Check(err)

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

	return bindAddress
}

func extract(installPath string, zipReader *zip.ReadCloser) (err error) {
	fmt.Print(".....")

	for i, file := range zipReader.File {
		fullname := path.Join(installPath, file.Name)
		fileInfo := file.FileInfo()
		if fileInfo.IsDir() {
			os.MkdirAll(fullname, fileInfo.Mode().Perm())
		} else {
			_, err := os.Stat(fullname)
			if err == nil {
				os.Remove(fullname)
			}

			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := fileInfo.Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return err
			}
			rc, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.CopyN(out, rc, fileInfo.Size())
			if err != nil {
				return err
			}
			rc.Close()
			out.Close()

			mtime := fileInfo.ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return err
			}

			err = files.CheckFilePermissions(installPath, file.Name)
			util.Check(err)

			if i%10 == 0 {
				fmt.Print(".")
			}
		}
	}
	return
}

type InstallerArgs struct {
	InstallPath string
	AdminGroup  string
	BindAddress string
	JoinServer  string
	JoinToken   string
}
