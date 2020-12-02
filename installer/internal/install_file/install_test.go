package install_file

import (
	"github.com/ruckstack/ruckstack/common/test_util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstallFile_Install(t *testing.T) {
	if testing.Short() {
		t.Skip("--short doesn't install files")
	}

	serverHome := test_util.TempPath("server-home-*")

	installFile, err := Parse(installerPackagePath)
	assert.NoError(t, err)

	err = installFile.Install(InstallOptions{
		AdminGroup:  test_util.GetCurrentUserGroup(t).Name,
		BindAddress: "127.0.0.1",
		JoinToken:   "none",
		TargetDir:   serverHome,
	})

	assert.NoError(t, err)
	//most files checked in extract_test.go
	assert.FileExists(t, serverHome+"/bin/example-manager")
	assert.FileExists(t, serverHome+"/config/cluster.config")
	assert.FileExists(t, serverHome+"/config/local.config")

}
