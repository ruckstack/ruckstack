package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

/**
System.Config file contains build-time system configuration settings.
*/
type SystemConfig struct {
	ManagerFilename string     `yaml:"managerFilename"`
	Proxy           []OpenPort `yaml:"proxy"`
}

type OpenPort struct {
	ServiceName string `yaml:"serviceName" validate:"required"`
	ServicePort int    `yaml:"servicePort" validate:"required"`
	Port        int    `yaml:"port" validate:"required"`
}

func ReadSystemConfig(content io.ReadCloser) (*SystemConfig, error) {
	systemConfig := new(SystemConfig)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(systemConfig); err != nil {
		return nil, fmt.Errorf("error parsing system.config: %s, ", err)
	}

	return systemConfig, nil
}

func LoadSystemConfig(serverHome string) (*SystemConfig, error) {
	file, err := os.Open(serverHome + "/config/system.config")
	if err != nil {
		return nil, err
	}

	return ReadSystemConfig(file)
}
