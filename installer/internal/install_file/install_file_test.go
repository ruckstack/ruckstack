package install_file

import (
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/stretchr/testify/assert"
	"testing"
)

var installerPackagePath = global_util.GetSourceRoot() + "/tmp/test-installer/out/example_1.0.5.installer"

func TestParse(t *testing.T) {
	installFile, err := Parse(installerPackagePath)
	assert.NoError(t, err)

	assert.Equal(t, installerPackagePath, installFile.FilePath)
	assert.NotNil(t, installFile.PackageConfig)
	assert.Equal(t, "Example Project", installFile.PackageConfig.Name)
}
