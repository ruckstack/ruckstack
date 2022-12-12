package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

type LocalConfigType struct {
	AdminGroup string `yaml:"adminGroup"`
	//BindAddress          string `yaml:"bindAddress"`
	//BindAddressInterface string // `yaml:"bindAddressInterface"`
	Join struct {
		Server string `yaml:"server"`
		Token  string `yaml:"token"`
	} `yaml:"join"`

	AdminGroupId int64
}

func ReadLocalConfig(content io.ReadCloser) (*LocalConfigType, error) {
	localConfig := new(LocalConfigType)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(localConfig); err != nil {
		return nil, fmt.Errorf("error parsing local.config: %s, ", err)
	}

	//ifaces, err := net.Interfaces()
	//if err != nil {
	//	return nil, err
	//}

	//for _, iface := range ifaces {
	//	addrs, err := iface.Addrs()
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	for _, addr := range addrs {
	//		var ip net.IP
	//		switch v := addr.(type) {
	//		case *net.IPNet:
	//			ip = v.IP
	//		case *net.IPAddr:
	//			ip = v.IP
	//		}
	//
	//		if ip.To4().String() == localConfig.BindAddress {
	//			localConfig.BindAddressInterface = iface.Name
	//		}
	//	}
	//}
	//
	//if localConfig.BindAddressInterface == "" {
	//	ui.Printf("WARNING: Cannot find network interface with IP %s", localConfig.BindAddress)
	//}

	return localConfig, nil
}

func LoadLocalConfig(serverHome string) (*LocalConfigType, error) {
	filePath := filepath.FromSlash(serverHome + "/config/local.config")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return ReadLocalConfig(file)
}

func (localConfig *LocalConfigType) Save() error {
	if err := os.MkdirAll(filepath.FromSlash(ServerHome+"/config"), 0755); err != nil {
		return err
	}

	localConfigFile, err := os.OpenFile(filepath.FromSlash(ServerHome+"/config/local.config"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	localConfigEncoder := yaml.NewEncoder(localConfigFile)
	if err := localConfigEncoder.Encode(localConfig); err != nil {
		return err
	}
	//if err := packageConfig.CheckFilePermissions(filepath.FromSlash("config/local.config"), localConfig, serverHome); err != nil {
	//	return err
	//}

	return nil
}
