package global_util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsRunningTests(t *testing.T) {
	assert.True(t, IsRunningTests())
}

func TestGetSourceRoot(t *testing.T) {
	assert.NotEqual(t, "", GetSourceRoot())
	assert.FileExists(t, GetSourceRoot()+"/LICENSE")

}
