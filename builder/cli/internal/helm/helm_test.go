package helm

import (
	"bytes"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func Test_init(t *testing.T) {
	assert.Contains(t, helmHome, "cache/helm")
	assert.DirExists(t, helmHome)
	assert.FileExists(t, helmHome+"/config/helm/repositories.yaml")
}

func TestReIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not download things from the internet")
	}

	stableChartsPath := environment.CachePath("helm/cache/helm/repository/stable-charts.txt")
	stableIndexPath := environment.CachePath("helm/cache/helm/repository/stable-index.yaml")

	assert.Nil(t, os.RemoveAll(environment.CachePath("helm/cache")))
	assert.Nil(t, ReIndex())
	assert.FileExists(t, stableChartsPath)
	assert.FileExists(t, stableIndexPath)

	stableChartsText, err := ioutil.ReadFile(stableChartsPath)
	assert.Nil(t, err)
	assert.Contains(t, string(stableChartsText), "postgresql")

	stableIndexText, err := ioutil.ReadFile(stableIndexPath)
	assert.Nil(t, err)
	assert.Contains(t, string(stableIndexText), "description: Chart for PostgreSQL, an object-relational database management system")
}

func TestSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not search the internet")
	}

	//ensure Search re-indexes when needed
	assert.Nil(t, os.RemoveAll(environment.CachePath("helm/cache")))

	output := new(bytes.Buffer)
	ui.SetOutput(output)
	defer ui.SetOutput(os.Stdout)

	type args struct {
		chartRepo string
		chartName string
	}
	tests := []struct {
		name           string
		args           args
		outputContains string
		wantErr        bool
	}{
		{
			name: "Find postgresql",
			args: args{
				chartName: "postgresql",
				chartRepo: "stable",
			},
			outputContains: "Chart: stable/postgresql",
		},
		{
			name: "Find coredns",
			args: args{
				chartName: "coredns",
				chartRepo: "stable",
			},
			outputContains: "Chart: stable/coredns",
		},
		{
			name: "Invalid chartName",
			args: args{
				chartName: "invalid",
				chartRepo: "stable",
			},
			wantErr:        true,
			outputContains: "unknown helm chart: stable/invalid",
		},
		{
			name: "Invalid repo",
			args: args{
				chartName: "postgresql",
				chartRepo: "invalid",
			},
			wantErr:        true,
			outputContains: "unknown helm repository: invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Search(tt.args.chartRepo, tt.args.chartName)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.outputContains)
			} else {
				assert.Nil(t, err)
				assert.Contains(t, output.String(), tt.outputContains)
				assert.Contains(t, output.String(), "All Available Versions:")
			}
		})
	}
}

func TestDownloadChart(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not download from the internet")
	}

	output := new(bytes.Buffer)
	ui.SetOutput(output)
	defer ui.SetOutput(os.Stdout)

	//ensure not already downloaded
	assert.Nil(t, os.RemoveAll(environment.CachePath("helm/download")))

	type args struct {
		repo    string
		chart   string
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Download postgresql",
			args: args{
				repo:    "stable",
				chart:   "postgresql",
				version: "8.1.2",
			},
			want: "asdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DownloadChart(tt.args.repo, tt.args.chart, tt.args.version)
			if tt.wantErr {
				assert.Equal(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
				assert.Contains(t, got, "/cache/helm/download/stable/postgresql-8.1.2.tgz")
				assert.FileExists(t, got)
				assert.Contains(t, output.String(), "Downloading chart")

				//does not re-download
				output.Reset()
				got, err = DownloadChart(tt.args.repo, tt.args.chart, tt.args.version)
				assert.Nil(t, err)
				assert.Contains(t, got, "/cache/helm/download/stable/postgresql-8.1.2.tgz")
				assert.NotContains(t, output.String(), "Downloading chart")
				assert.Contains(t, output.String(), "Already downloaded")

			}
		})
	}
}
