package cmd

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/helm"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/spf13/cobra"
)

var (
	chartName string
	chartRepo string
)

var helmSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Simple Helm interface",
	Run: func(cmd *cobra.Command, args []string) {
		helm.Setup()
		helm.Search(chartRepo, chartName)
	},
}

func init() {
	helmSearchCmd.Flags().StringVar(&chartName, "chart", "", "Chart to search (required)")
	helmSearchCmd.Flags().StringVar(&chartRepo, "repo", "stable", "Chart repository to search. Defaults to 'stable'")

	util.Check(helmSearchCmd.MarkFlagRequired("chart"))

	helmRootCmd.AddCommand(helmSearchCmd)

}
