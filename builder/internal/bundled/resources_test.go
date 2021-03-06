package bundled

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
)

func TestOpenFile(t *testing.T) {
	initFile, err := OpenFile("init/empty/ruckstack.yaml")
	if assert.NoError(t, err) {
		assert.NotNil(t, initFile)
	}
}

func TestOpenDir(t *testing.T) {
	dir, err := OpenDir("init")
	if assert.NoError(t, err) {
		err = fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
			fmt.Println(path)

			return nil
		})
		assert.NoError(t, err)
	}
}
