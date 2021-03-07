package global_util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestTarAndUntar(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not work with large files")
	}

	type args struct {
		sourceDir      string
		targetFilename string
		compress       bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "compress dir",
			args: args{
				sourceDir:      filepath.Join(GetSourceRoot(), "builder", "internal", "bundled", "install_dir", "data"),
				targetFilename: filepath.Join(GetSourceRoot(), "tmp", "tar_test", "compress_dir.tgz"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.RemoveAll(filepath.Join(GetSourceRoot(), "tmp", "tar_test"))

			_ = os.MkdirAll(filepath.Dir(tt.args.targetFilename), 0755)
			err := TarDirectory(tt.args.sourceDir, tt.args.targetFilename, tt.args.compress)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					uncompressedPath := filepath.Join(GetSourceRoot(), "tmp", "tar_test", "uncompressed")
					_ = os.MkdirAll(uncompressedPath, 0755)

					err := UntarFile(tt.args.targetFilename, uncompressedPath, tt.args.compress)
					if assert.NoError(t, err) {
						assert.FileExists(t, filepath.Join(uncompressedPath, "kubectl", "README.txt"))
					}

				}
			}
		})
	}
}
