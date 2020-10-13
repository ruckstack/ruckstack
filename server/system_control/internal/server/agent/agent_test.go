package agent

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStart(t *testing.T) {
	assert.NoError(t, Start())
}
