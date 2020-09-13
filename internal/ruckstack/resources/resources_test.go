package resources

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetResourcePath(t *testing.T) {
	path, err := ResourcePath("new_project")
	assert.Nil(t, err)
	assert.DirExists(t, path)
	assert.Regexp(t, ".*/resources/.*", path, "resources")

	path, err = ResourcePath("new_project", "example", "ruckstack.conf")
	assert.Nil(t, err)
	assert.FileExists(t, path)

	path, err = ResourcePath("invalid")
	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}
