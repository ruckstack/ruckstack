package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type ClusterConfigType struct {
	DevModeEnabled bool
}

func ReadClusterConfig(content io.ReadCloser) (*ClusterConfigType, error) {
	clusterConfig := new(ClusterConfigType)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(clusterConfig); err != nil {
		return nil, fmt.Errorf("error parsing cluster.config: %s, ", err)
	}

	return clusterConfig, nil
}

func LoadClusterConfig(serverHome string) (*ClusterConfigType, error) {
	file, err := os.Open(serverHome + "/config/cluster.config")
	if err != nil {
		return nil, err
	}

	return ReadClusterConfig(file)
}

func (clusterConfig *ClusterConfigType) Save() error {
	if err := os.MkdirAll(ServerHome+"/config", 0755); err != nil {
		return err
	}

	clusterConfigFile, err := os.OpenFile(ServerHome+"/config/cluster.config", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	clusterConfigEncoder := yaml.NewEncoder(clusterConfigFile)
	if err := clusterConfigEncoder.Encode(clusterConfig); err != nil {
		return err
	}
	//if err := packageConfig.CheckFilePermissions("config/cluster.config", localConfig, serverHome); err != nil {
	//	return err
	//}

	return nil
}
