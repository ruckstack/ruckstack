package k3s

import (
	"context"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStart(t *testing.T) {
	if !environment.IsRunningAsRoot {
		t.Skip("Cannot run TestStart as non-root user")
	}

	assert.NoError(t, Start(context.Background()))
}
