package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRuckstackHome(t *testing.T) {
	home := GetRuckstackHome()
	assert.Regexp(t, "github.com/ruckstack/ruckstack$", home)
}

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test", "path")
	assert.Contains(t, tempDir, "test")
	assert.Contains(t, tempDir, "path")

}
