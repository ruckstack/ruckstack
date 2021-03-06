package util

import (
	"github.com/ruckstack/ruckstack/builder/internal/bundled"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not download files")
	}
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can download a file",
			args: args{
				url: "https://example.com/index.html",
			},
		},
		{
			name:    "Error on invalid file",
			wantErr: "cannot download https://example.com/invalid: 404 Not Found",
			args: args{
				url: "https://example.com/invalid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := DownloadFile(tt.args.url)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.FileExists(t, path)

				file, err := ioutil.ReadFile(path)
				assert.NoError(t, err)
				assert.Contains(t, string(file), "Example Domain")

			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}

func TestExtractFromGzip(t *testing.T) {
	type args struct {
		gzipSource string
		wantedFile string
	}
	tests := []struct {
		name                string
		args                args
		expectedFileContent string
		wantErr             string
	}{
		{
			name:                "Can extract root file",
			expectedFileContent: "This is the root file",
			args: args{
				gzipSource: "util_test.tar.gz",
				wantedFile: "root_file.txt",
			},
		},
		{
			name:                "Can extract nested file",
			expectedFileContent: "Subdir 1 File 2",
			args: args{
				gzipSource: "util_test.tar.gz",
				wantedFile: "subdir1/subdir1_file_2.txt",
			},
		},
		{
			name:    "Error if gzip file does not exist",
			wantErr: "cannot open file to extract: open invalid.tar.gz: no such file or directory",
			args: args{
				gzipSource: "invalid.tar.gz",
				wantedFile: "root_file.txt",
			},
		},
		{
			name:    "Error if requested file does not exist",
			wantErr: "cannot find invalid.txt in util_test.tar.gz",
			args: args{
				gzipSource: "util_test.tar.gz",
				wantedFile: "invalid.txt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractedFile, err := ExtractFromGzip(tt.args.gzipSource, tt.args.wantedFile)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.FileExists(t, extractedFile)

				fileContent, err := ioutil.ReadFile(extractedFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFileContent, strings.TrimSpace(string(fileContent)))
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, tt.wantErr, err.Error())
				}

			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can copy file",
			args: args{
				source: "util.go",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetFile := environment.TempPath("util_test_copy-*.dst")
			sourceFile, err := os.DirFS(".").Open(tt.args.source)
			if !assert.NoError(t, err) {
				return
			}

			err = CopyFile(sourceFile, targetFile)

			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.FileExists(t, targetFile)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				assert.NoFileExists(t, targetFile)
			}
		})
	}
}

func TestCopyDir(t *testing.T) {
	type args struct {
		source fs.FS
	}
	newProjectPath, err := bundled.OpenDir("init")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can copy files",
			args: args{
				source: newProjectPath,
			},
		},
		{
			name:    "Invalid source",
			wantErr: "open invalid.dir/.: no such file or directory",
			args: args{
				source: os.DirFS("invalid.dir"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetDir := environment.TempPath("util_test_copy_dir-*")

			err := CopyDir(tt.args.source, targetDir)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.FileExists(t, filepath.Join(targetDir, "empty/ruckstack.yaml"))
				assert.FileExists(t, filepath.Join(targetDir, "example/ruckstack.yaml"))
				assert.FileExists(t, filepath.Join(targetDir, "example/frontend/Dockerfile"))
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}

		})
	}
}
