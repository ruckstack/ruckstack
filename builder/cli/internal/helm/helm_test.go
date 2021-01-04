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
	assert.FileExists(t, helmHome+"/repositories.yaml")
}

func TestReIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not download things from the internet")
	}

	stableChartsPath := environment.CachePath("helm/cache/helm/repository/stable-charts.txt")
	stableIndexPath := environment.CachePath("helm/cache/helm/repository/stable-index.yaml")

	assert.Nil(t, ReIndex())
	assert.FileExists(t, stableChartsPath)
	assert.FileExists(t, stableIndexPath)

	stableChartsText, err := ioutil.ReadFile(stableChartsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(stableChartsText), "postgresql")

	stableIndexText, err := ioutil.ReadFile(stableIndexPath)
	assert.NoError(t, err)
	assert.Contains(t, string(stableIndexText), "description: Chart for PostgreSQL, an object-relational database management system")
}

func TestDownloadChart(t *testing.T) {
	if testing.Short() {
		t.Skip("-short tests do not download from the internet")
	}

	ui.SetVerbose(true)
	defer ui.SetVerbose(false)

	output := new(bytes.Buffer)
	ui.SetOutput(output)

	type args struct {
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
				chart:   "stable/postgresql",
				version: "8.1.2",
			},
			want: "asdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//delete it for the original download
			_ = os.Remove(environment.CachePath("/download/helm/stable/postgresql-8.1.2.tgz"))

			got, err := DownloadChart(tt.args.chart, tt.args.version)
			if tt.wantErr {
				assert.Equal(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, "/cache/download/helm/stable/postgresql-8.1.2.tgz")
				assert.FileExists(t, got)

				//does not re-download
				output.Reset()
				got, err = DownloadChart(tt.args.chart, tt.args.version)
				assert.NoError(t, err)
				assert.Contains(t, got, "/cache/download/helm/stable/postgresql-8.1.2.tgz")
				assert.Contains(t, output.String(), "Already downloaded")

			}
		})
	}
}
