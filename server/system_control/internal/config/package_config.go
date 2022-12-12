package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

type PackageConfigType struct {
	Id        string
	Name      string
	Version   string
	BuildTime int64 `yaml:"buildTime"`
}

func ReadPackageConfig(content io.ReadCloser) (*PackageConfigType, error) {
	packageConfig := new(PackageConfigType)

	decoder := yaml.NewDecoder(content)

	if err := decoder.Decode(packageConfig); err != nil {
		return nil, fmt.Errorf("error parsing package.config: %s, ", err)
	}

	return packageConfig, nil
}

func LoadPackageConfig(serverHome string) (*PackageConfigType, error) {
	file, err := os.Open(filepath.FromSlash(serverHome + "/.package.config"))
	if err != nil {
		return nil, err
	}

	return ReadPackageConfig(file)
}
