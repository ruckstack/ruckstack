package resources

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestGetResourcePath(t *testing.T) {
	path, err := ResourcePath("new_project", "example", "ruckstack.conf")
	assert.Nil(t, err)
	assert.True(t, strings.Contains(path, "resources"))

	stat, err := os.Stat(path)
	assert.Nil(t, err)
	assert.True(t, !stat.IsDir())
}
