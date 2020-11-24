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
					ManifestServices: []service.ManifestService{
						{},
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
					ManifestServices: []service.ManifestService{
						{
							Id:             "service-id",
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

func TestProject_GetServices(t *testing.T) {

	project := Project{
		ManifestServices: []service.ManifestService{
			{
				Id: "manifest-1",
			},
			{
				Id: "manifest-2",
			},
		},
		HelmServices: []service.HelmService{
			{
				Id: "helm-1",
			},
			{
				Id: "helm-2",
			},
		},
		DockerfileServices: []service.DockerfileService{
			{
				Id: "docker-1",
			},
			{
				Id: "docker-2",
			},
		},
	}
	assert.Equal(t, 6, len(project.GetServices()))
	assert.Equal(t, "manifest-1", project.GetServices()[0].GetId())
	assert.Equal(t, "manifest-2", project.GetServices()[1].GetId())
	assert.Equal(t, "helm-1", project.GetServices()[2].GetId())
	assert.Equal(t, "helm-2", project.GetServices()[3].GetId())
	assert.Equal(t, "docker-1", project.GetServices()[4].GetId())
	assert.Equal(t, "docker-2", project.GetServices()[5].GetId())

}
