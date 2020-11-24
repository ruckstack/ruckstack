package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
	"net/url"
)

func init() {
	var internalBuildCommand = &cobra.Command{
		Use:    "internal-build",
		Short:  "Commands used in internal build process",
		Hidden: true,
	}

	initDownload(internalBuildCommand)

	RootCmd.AddCommand(internalBuildCommand)
}

func initDownload(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "download",
		Short: "Downloads a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, urlToDownload := range []string{
				fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s-airgap-images-amd64.tar", url.PathEscape(environment.PackagedK3sVersion)),
				fmt.Sprintf("https://github.com/rancher/k3s/releases/download/v%s/k3s", url.PathEscape(environment.PackagedK3sVersion)),
				fmt.Sprintf("https://get.helm.sh/helm-v%s-linux-amd64.tar.gz", url.PathEscape(environment.PackagedHelmVersion)),
			} {
				ui.Printf("Downloading %s...", urlToDownload)
				_, err := util.DownloadFile(urlToDownload)
				if err != nil {
					return err
				}

			}

			return nil
		},
	}

	parent.AddCommand(cmd)
}
