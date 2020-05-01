package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/spf13/cobra"
)

var kubectlCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Kubernetes CLI",
	Long: `Kubectl is a low-level Kubernetes command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		k3s.ExecKubectl(args...)
	},
}

func init() {
	systemCmd.AddCommand(kubectlCmd)
}
