package init_project

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestNewProject_example(t *testing.T) {
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	environment.OutDir = environment.TempPath("test_new_example_project-*")

	err := InitProject("example")
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(environment.OutDir, "ruckstack.yaml"))
	assert.FileExists(t, filepath.Join(environment.OutDir, "backend", "Dockerfile"))
	assert.FileExists(t, filepath.Join(environment.OutDir, "backend", "server.js"))
	assert.FileExists(t, filepath.Join(environment.OutDir, "ruckstack", "README.txt"))

	assert.Contains(t, output.String(), "Created example project in")
}

func TestNewProject_starter(t *testing.T) {
	environment.OutDir = environment.TempPath("test_new_starter_project-*")

	err := InitProject("empty")
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(environment.OutDir, "ruckstack.yaml"))
	assert.NoFileExists(t, filepath.Join(environment.OutDir, "cart"))
}

func TestNewProject_invalid(t *testing.T) {
	environment.OutDir = environment.TempPath("test_new_invalid_project-*")

	err := InitProject("invalid")
	assert.Equal(t, err.Error(), "unknown template: 'invalid'. Available templates: empty, example, wordpress")
}
