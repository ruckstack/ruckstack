package util

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetRuckstackHome(t *testing.T) {
	home := GetRuckstackHome()
	assert.True(t, strings.HasSuffix(strings.ReplaceAll(home, "\\", "/"), "github.com/ruckstack/ruckstack"))
}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test", "path")
	assert.Contains(t, tempDir, "test")
	assert.Contains(t, tempDir, "path")

}
