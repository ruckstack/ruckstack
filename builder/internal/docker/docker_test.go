package docker

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	alpineImage = "alpine:3.12"
)

func TestImagePull(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not pull docker images")
	}
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	type args struct {
		imageRef string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can pull image",
			args: args{
				imageRef: alpineImage,
			},
		},
		{
			name:    "Error on invalid image",
			wantErr: true,
			args: args{
				imageRef: "invalid:3.12",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ImagePull(tt.args.imageRef)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output.String(), "Pulling from library/alpine")
			}

		})
	}
}

func TestSaveImages(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	outputPath := environment.TempPath("test_docker_save-*.tar")
	type args struct {
		outputPath string
		imageRefs  []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Can save image",
			args: args{
				outputPath: outputPath,
				imageRefs:  []string{"alpine:3.12"},
			},
		},
		{
			name:    "Error on invalid image",
			wantErr: "Error saving images alpine:3.12, invalid:1.3: reference does not exist",
			args: args{
				outputPath: outputPath,
				imageRefs:  []string{"alpine:3.12", "invalid:1.3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveImages(tt.args.outputPath, tt.args.imageRefs...)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.FileExists(t, outputPath)
			}
		})
	}
	_ = os.Remove(outputPath)
}

func TestImageBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not run docker images")
	}
	output := new(bytes.Buffer)
	ui.SetOutput(output)

	dockerfile := global_util.GetSourceRoot() + "/Dockerfile"

	type args struct {
		dockerfile string
		tags       []string
		labels     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Can build",
			args: args{
				dockerfile: dockerfile,
				tags:       []string{"ruckstack:test-built"},
				labels:     map[string]string{"ruckstack.test": "true"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ImageBuild(tt.args.dockerfile, tt.args.tags, tt.args.labels)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output.String(), "Successfully tagged ruckstack:test-built")
			}

		})
	}
}
