package containerd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not start containerd")
	}

	err := Start()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	serving, err := client.IsServing(context.Background())
	assert.NoError(t, err)
	assert.True(t, serving)
}

func Test_LoadPackagedImages(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not start containerd")
	}

	assert.NoError(t, Start())

	assert.NoError(t, LoadPackagedImages())
}
