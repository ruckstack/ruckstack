package config

type AddNodeToken struct {
	Token      string `yaml:"token"`
	Server     string `yaml:"server"`
	KubeConfig string `yaml:"kubeConfig"`
}
