package environment

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	assert.NotEqual(t, "/", ServerHome)
	assert.FileExists(t, ServerHome+"/.package.config")
}
