package environment

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	assert.NotEqual(t, "/", ruckstackHome)
	assert.FileExists(t, ruckstackHome+"/LICENSE")

	assert.Equal(t, resourceRoot, ruckstackHome+"/builder/cli/install_root/resources")
	assert.DirExists(t, resourceRoot)

	assert.Equal(t, cacheRoot, ruckstackHome+"/cache")
	assert.DirExists(t, cacheRoot)

}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test/path")
	assert.Contains(t, tempDir, "/test/path")
}

func TestGetResourcePath(t *testing.T) {
	path, err := ResourcePath("new_project")
	assert.Nil(t, err)
	assert.DirExists(t, path)
	assert.Regexp(t, ".*/resources/.*", path, "resources")

	path, err = ResourcePath("new_project/example/ruckstack.conf")
	assert.Nil(t, err)
	assert.FileExists(t, path)

	path, err = ResourcePath("invalid")
	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}
