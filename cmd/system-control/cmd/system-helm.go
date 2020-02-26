package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/helm"
	"github.com/spf13/cobra"
)

var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Helm CLI",
	Long: `helm is a low-level Kubernetes command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		helm.ExecHelm(args...)
	},
}

func init() {
	systemCmd.AddCommand(helmCmd)
}
