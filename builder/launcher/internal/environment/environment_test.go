package environment

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTemPath(t *testing.T) {
	tempDir := TempPath("test/path")
	assert.Contains(t, tempDir, "/test/path")
}
