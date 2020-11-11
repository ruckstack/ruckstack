package containerd

import (
	"context"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/k3s"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_LoadPackagedImages(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not start containerd")
	}

	assert.NoError(t, k3s.Start(context.Background()))

	assert.NoError(t, LoadPackagedImages())
}
