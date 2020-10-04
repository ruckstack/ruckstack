package install_file

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/test_util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
	"os"
	"testing"
)

func TestExtract(t *testing.T) {
	if testing.Short() {
		t.Skip("--short doesn't extract files")
	}

	serverHome := test_util.TempPath("server-home-*")

	currentUserGroup := test_util.GetCurrentUserGroup(t)

	localConfig := &config.LocalConfig{
		AdminGroup:  currentUserGroup.Name,
		BindAddress: "1.2.3.4",
	}

	installFile, err := Parse(installerPackagePath)
	assert.NoError(t, err)

	err = installFile.Extract(serverHome, localConfig)
	if assert.NoError(t, err) {
		assert.FileExists(t, serverHome+"/.package.config")
		assert.FileExists(t, serverHome+"/bin/system-control")
		assert.NoFileExists(t, serverHome+"/config/cluster.config") //created by install, not packaged
		assert.NoFileExists(t, serverHome+"/config/local.config")   //created by install, not packaged
		assert.FileExists(t, serverHome+"/lib/helm")
		assert.FileExists(t, serverHome+"/lib/k3s")
		assert.FileExists(t, serverHome+"/data/server/manifests/traefik.yaml")
		assert.FileExists(t, serverHome+"/data/server/static/charts/cart.tgz")
		assert.FileExists(t, serverHome+"/data/web/site-down.html")

		assert.NoFileExists(t, serverHome+"/data/agent/images/images.tar")
		assert.DirExists(t, serverHome+"/data/agent/images/images.untar")

		assert.NoFileExists(t, serverHome+"/data/agent/images/k3s.tar")
		assert.DirExists(t, serverHome+"/data/agent/images/k3s.untar")

		for _, file := range []string{"/bin/system-control", "/lib/helm", "/lib/k3s"} {
			file = serverHome + file
			stat, err := os.Stat(file)
			assert.NoError(t, err)
			assert.True(t, stat.Mode()&(unix.R_OK*100) != 0, fmt.Sprintf("%s is not owner-readablele", file))
			assert.True(t, stat.Mode()&(unix.W_OK*100) != 0, fmt.Sprintf("%s is not owner-writable", file))
			assert.True(t, stat.Mode()&(unix.X_OK*100) != 0, fmt.Sprintf("%s is not owner-executable", file))

			assert.True(t, stat.Mode()&(unix.R_OK*10) != 0, fmt.Sprintf("%s is not group-readablele", file))
			assert.False(t, stat.Mode()&(unix.W_OK*10) != 0, fmt.Sprintf("%s should not be not group-writable", file))
			assert.True(t, stat.Mode()&(unix.X_OK*100) != 0, fmt.Sprintf("%s is not owner-executable", file))

			assert.False(t, stat.Mode()&0001 != 0, fmt.Sprintf("%s should not be other-executable", file))

		}
	}
}
