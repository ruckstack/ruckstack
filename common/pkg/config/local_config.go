package config

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
)

type LocalConfig struct {
	AdminGroup           string `yaml:"adminGroup"`
	BindAddress          string `yaml:"bindAddress"`
	BindAddressInterface string // `yaml:"bindAddressInterface"`
	Join                 struct {
		Server string `yaml:"server"`
		Token  string `yaml:"token"`
	} `yaml:"join"`

	AdminGroupId int64
}

func ReadLocalConfig(content io.ReadCloser) (*LocalConfig, error) {
	localConfig := new(LocalConfig)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(localConfig); err != nil {
		return nil, fmt.Errorf("error parsing local.config: %s, ", err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.To4().String() == localConfig.BindAddress {
				localConfig.BindAddressInterface = iface.Name
			}
		}
	}

	if localConfig.BindAddressInterface == "" {
		ui.Printf("WARNING: Cannot find network interface with IP %s", localConfig.BindAddress)
	}

	return localConfig, nil
}

func LoadLocalConfig(serverHome string) (*LocalConfig, error) {
	file, err := os.Open(serverHome + "/config/local.config")
	if err != nil {
		return nil, err
	}

	return ReadLocalConfig(file)
}

func (localConfig *LocalConfig) Save(serverHome string, packageConfig *PackageConfig) error {
	if err := os.MkdirAll(serverHome+"/config", 0755); err != nil {
		return err
	}

	localConfigFile, err := os.OpenFile(serverHome+"/config/local.config", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	localConfigEncoder := yaml.NewEncoder(localConfigFile)
	if err := localConfigEncoder.Encode(localConfig); err != nil {
		return err
	}
	if err := packageConfig.CheckFilePermissions("config/local.config", localConfig, serverHome); err != nil {
		return err
	}

	return nil
}
