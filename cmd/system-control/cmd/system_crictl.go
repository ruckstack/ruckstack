package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/spf13/cobra"
)

var crictlCmd = &cobra.Command{
	Use: "crictl",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Short: "CRI CLI",
	Long: `Crictl is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		k3s.ExecCrictl(args...)
	},
}

func init() {
	systemCmd.AddCommand(crictlCmd)
}
