package settings

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type SettingsConfig struct {
	AskedAnalytics bool
	AllowAnalytics bool
	AnalyticsId    string
	LicenseKey     string
}

var Settings *SettingsConfig
var settingsFilePath = filepath.Join(environment.RuckstackHome, "settings.yaml")

func init() {
	Settings = &SettingsConfig{}

	_, err := os.Stat(settingsFilePath)
	if err == nil {
		settingsFile, err := os.OpenFile(settingsFilePath, os.O_RDONLY, 0644)
		if err != nil {
			ui.Fatalf("Cannot open settings file: %s", err)
		}
		defer settingsFile.Close()

		decoder := yaml.NewDecoder(settingsFile)
		if err := decoder.Decode(Settings); err != nil {
			ui.Fatalf("Cannot parse settings file %s: %s", settingsFilePath, err)
		}
	}
}

func (config *SettingsConfig) Save() error {
	settingsFile, err := os.OpenFile(settingsFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open settings file: %s", err)
	}
	defer settingsFile.Close()

	encoder := yaml.NewEncoder(settingsFile)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error saving settings: %s", err)
	}

	return nil
}
