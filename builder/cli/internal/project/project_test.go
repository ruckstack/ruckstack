package project

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/project/service"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProject_Validate(t *testing.T) {
	type args struct {
		project *Project
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name:    "Empty project fails validation",
			wantErr: "error parsing project file: Key: 'Project.Id' Error:Field validation for 'Id' failed on the 'required' tag",
			args: args{
				project: &Project{},
			},
		},
		{
			name:    "Requires at least one service",
			wantErr: "error parsing project file: at least one service block is required",
			args: args{
				project: &Project{
					Id:      "test-project",
					Name:    "Test Project",
					Version: "1.2.3",
				},
			},
		},
		{
			name:    "Empty services fail validation",
			wantErr: "error parsing service : Key: 'ManifestService.Id' Error:Field validation for 'Id' failed on the 'required' tag",
			args: args{
				project: &Project{
					Id:      "test-project",
					Name:    "Test Project",
					Version: "1.2.3",
					Services: []Service{
						&service.ManifestService{},
					},
				},
			},
		},
		{
			name: "Minimum project passes validation",
			args: args{
				project: &Project{
					Id:      "test-project",
					Name:    "Test Project",
					Version: "1.2.3",
					Services: []Service{
						&service.ManifestService{
							Id:             "service-id",
							Type:           "manifest",
							Port:           1234,
							ProjectId:      "project-id",
							ProjectVersion: "1.2.3",
							Manifest:       "test-manifest.yaml",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.project.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.True(t, strings.HasPrefix(err.Error(), tt.wantErr), err)
				}

			}
		})
	}
}
