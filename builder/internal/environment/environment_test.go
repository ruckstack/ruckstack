package environment

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	assert.NotEqual(t, "/", RuckstackHome)
	assert.FileExists(t, RuckstackHome+"/LICENSE")

	assert.Equal(t, resourceRoot, RuckstackHome+"/builder/cli/install_root/resources")
	assert.DirExists(t, resourceRoot)

	assert.Equal(t, cacheRoot, RuckstackHome+"/cache")
	assert.DirExists(t, cacheRoot)

}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test/path")
	assert.Contains(t, tempDir, "/test/path")
}

func TestGetResourcePath(t *testing.T) {
	path, err := ResourcePath("new_project")
	assert.NoError(t, err)
	assert.DirExists(t, path)
	assert.Regexp(t, ".*/resources/.*", path, "resources")

	path, err = ResourcePath("new_project/example/ruckstack.conf")
	assert.NoError(t, err)
	assert.FileExists(t, path)

	path, err = ResourcePath("invalid")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
