package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/spf13/cobra"
	"os"
)

var ctrCmd = &cobra.Command{
	Use: "ctr",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Short: "Containerd CLI",
	Long: `Ctr is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		k3s.ExecCtr(os.Stdout, os.Stderr, args...)
	},
}

func init() {
	systemCmd.AddCommand(ctrCmd)
}
