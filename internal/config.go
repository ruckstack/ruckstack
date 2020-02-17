package internal

type PackageConfig struct {
	Id                string
	Name              string
	Version           string
	BuildTime         int64  `yaml:"buildTime"`
	SystemControlName string `yaml:"systemControlName"`

	FilePermissions map[string]InstalledFileConfig `yaml:"filePermissions"`
	Files           map[string]string
}

type SystemConfig struct {
}

type LocalConfig struct {
	AdminGroup  string `yaml:"adminGroup"`
	BindAddress string `yaml:"bindAddressps"`
}

type InstalledFileConfig struct {
	AdminGroupReadable bool `yaml:"adminGroupReadable"`
	AdminGroupWritable bool `yaml:"adminGroupWritable"`
	Executable         bool
}