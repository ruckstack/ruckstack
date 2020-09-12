package new_project

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProject_example(t *testing.T) {
	outDir := util.TempPath("test_new_example_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "example")
	assert.Nil(t, err)
	assert.FileExists(t, filepath.Join(outDir, "ruckstack.conf"))
	assert.FileExists(t, filepath.Join(outDir, "cart", "Dockerfile"))
	assert.FileExists(t, filepath.Join(outDir, "homepage", "src", "index.jsp"))
}

func TestNewProject_starter(t *testing.T) {
	outDir := util.TempPath("test_new_starter_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "starter")
	assert.Nil(t, err)
	assert.FileExists(t, filepath.Join(outDir, "ruckstack.conf"))
	assert.NoFileExists(t, filepath.Join(outDir, "cart"))
}

func TestNewProject_invalid(t *testing.T) {
	outDir := util.TempPath("test_new_invalid_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "invalid")
	assert.Equal(t, err.Error(), "unknown template: 'invalid'. Available templates: example, starter")
}
