package global_util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_sha1File(t *testing.T) {
	hash, err := Sha1File("../../LICENSE")
	assert.NoError(t, err)
	assert.Equal(t, "71bf517eebffe4bca12a7ea41f30b8685412092e", hash)
}
