package service

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/install_file"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestService_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("--short does not build installers")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	type args struct {
		manifestFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can build yaml",
			args: args{
				manifestFile: "manifest_test_complete.yaml",
			},
		},
		{
			name:    "Fails on empty file",
			wantErr: "empty manifest file test-manifest.yaml",
			args: args{
				manifestFile: "manifest_test_empty.yaml",
			},
		},
		{
			name:    "Fails on invalid yaml",
			wantErr: "error parsing manifest test-manifest.yaml: yaml syntax error: did not find expected key",
			args: args{
				manifestFile: "manifest_test_invalid_yaml.yaml",
			},
		},
		{
			name:    "Fails on invalid manifest",
			wantErr: "error parsing manifest test-manifest.yaml: Object 'Kind' is missing in 'but: not a valid manifest\nthis: is valid yaml\n'",
			args: args{
				manifestFile: "manifest_test_invalid_manifest.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := environment.TempPath("manifest-test-*")
			environment.ProjectDir = filepath.Join(testDir, "base-dir")
			outFile := filepath.Join(testDir, "out.installer")

			assert.NoError(t, os.MkdirAll(environment.ProjectDir, 0755))

			manifestContent, err := ioutil.ReadFile(tt.args.manifestFile)
			assert.NoError(t, err)
			assert.NoError(t, ioutil.WriteFile(filepath.Join(environment.ProjectDir, "test-manifest.yaml"), manifestContent, 0644))

			service := &ManifestService{
				Id:             "test-service",
				Type:           "manifest",
				Port:           8000,
				ProjectId:      "test-project",
				ProjectVersion: "1.2.3",
				Manifest:       "test-manifest.yaml",
			}

			installFile, err := install_file.StartCreation(outFile)
			assert.NoError(t, err)

			err = service.Build(installFile)

			assert.NoError(t, installFile.CompleteCreation())

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.NoError(t, err)

				unzipPath := testDir + "/manifest_test_unzip"
				assert.NoError(t, global_util.UnzipFile(outFile, unzipPath))

				assert.FileExists(t, unzipPath+"/data/server/manifests/test-service.yaml")

				savedContents, err := ioutil.ReadFile(unzipPath + "/data/server/manifests/test-service.yaml")
				assert.NoError(t, err)
				assert.Contains(t, string(savedContents), "name: test-ds")

				manifestContent, err := ioutil.ReadFile(filepath.Join(unzipPath, "data/agent/images/images.untar/manifest.json"))
				assert.NoError(t, err)
				assert.Contains(t, string(manifestContent), "\"RepoTags\":[\"traefik:2.3\"]")

			}
		})
	}
}
