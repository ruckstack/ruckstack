package new_project

import (
	"bytes"
	"github.com/ruckstack/ruckstack/common"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProject_example(t *testing.T) {
	output := new(bytes.Buffer)
	ui.SetOutput(output)
	defer ui.SetOutput(os.Stdout)

	outDir := common.TempPath("test_new_example_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "example")
	assert.Nil(t, err)
	assert.FileExists(t, filepath.Join(outDir, "ruckstack.conf"))
	assert.FileExists(t, filepath.Join(outDir, "cart", "Dockerfile"))
	assert.FileExists(t, filepath.Join(outDir, "homepage", "src", "index.jsp"))

	assert.Contains(t, output.String(), "Created example project in")
}

func TestNewProject_starter(t *testing.T) {
	outDir := common.TempPath("test_new_starter_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "empty")
	assert.Nil(t, err)
	assert.FileExists(t, filepath.Join(outDir, "ruckstack.conf"))
	assert.NoFileExists(t, filepath.Join(outDir, "cart"))
}

func TestNewProject_invalid(t *testing.T) {
	outDir := common.TempPath("test_new_invalid_project")

	assert.Nil(t, os.RemoveAll(outDir))

	err := NewProject(outDir, "invalid")
	assert.Equal(t, err.Error(), "unknown template: 'invalid'. Available templates: empty, example")
}
