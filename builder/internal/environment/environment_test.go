package environment

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRuckstackHome(t *testing.T) {
	home := GetRuckstackHome()
	assert.NotEqual(t, "/", home)
	assert.FileExists(t, home+"/LICENSE")
}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test/path")
	assert.Contains(t, tempDir, "/test/path")
}
