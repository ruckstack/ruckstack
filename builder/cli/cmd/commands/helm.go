package commands

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/helm"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var helmCommand = &cobra.Command{
		Use:   "helm",
		Short: "Commands for interacting with the Helm repository",
	}

	initReIndex(helmCommand)
	initSearch(helmCommand)

	rootCmd.AddCommand(helmCommand)
}

func initReIndex(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "re-index",
		Short: "Refreshes list of available helm charts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helm.ReIndex(); err != nil {
				return err
			}

			return nil
		},
	}

	parent.AddCommand(cmd)
}

func initSearch(parent *cobra.Command) {
	var chartName string
	var chartRepo string

	var cmd = &cobra.Command{
		Use:   "search",
		Short: "Simple Helm search",
		RunE: func(cmd *cobra.Command, args []string) error {
			return helm.Search(chartRepo, chartName)
		},
	}

	cmd.Flags().StringVar(&chartName, "chart", "", "Chart to search")
	cmd.Flags().StringVar(&chartRepo, "repo", "stable", "Chart repository to search. Defaults to 'stable'")

	ui.MarkFlagsRequired(cmd, "chart")

	parent.AddCommand(cmd)
}
