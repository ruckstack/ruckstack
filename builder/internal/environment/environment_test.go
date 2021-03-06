package environment

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	assert.NotEqual(t, "/", RuckstackWorkDir)
	assert.DirExists(t, RuckstackWorkDir+"/cache")
	assert.DirExists(t, RuckstackWorkDir+"/tmp")
}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test/path")
	assert.Contains(t, tempDir, "/test/path")
}
