package config

type PackageConfig struct {
	Id                string
	Name              string
	Version           string
	BuildTime         int64  `yaml:"buildTime"`
	SystemControlName string `yaml:"systemControlName"`

	FilePermissions map[string]PackagedFileConfig `yaml:"filePermissions"`
	Files           map[string]string
}

type SystemConfig struct {
}

type LocalConfig struct {
	AdminGroup  string `yaml:"adminGroup"`
	BindAddress string `yaml:"bindAddress"`
	Join        struct {
		Server string `yaml:"server"`
		Token  string `yaml:"token"`
	} `yaml:"join"`
}

type PackagedFileConfig struct {
	AdminGroupReadable bool `yaml:"adminGroupReadable"`
	AdminGroupWritable bool `yaml:"adminGroupWritable"`
	Executable         bool
}

type AddNodeToken struct {
	Token      string `yaml:"token"`
	Server     string `yaml:"server"`
	KubeConfig string `yaml:"kubeConfig"`
}
