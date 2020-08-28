package commands

import (
	"github.com/ruckstack/ruckstack/internal/system-control/helm"
	"github.com/ruckstack/ruckstack/internal/system-control/k3s"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var systemCmd = &cobra.Command{
		Use:   "system",
		Short: "Low-Level System Commands",
	}

	initSystemCrictl(systemCmd)
	initSystemCtr(systemCmd)
	initSystemKubectl(systemCmd)
	initSystemHelm(systemCmd)

	rootCmd.AddCommand(systemCmd)
}

func initSystemCrictl(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use: "crictl",
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		Short: "CRI CLI",
		Long: `Crictl is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			k3s.ExecCrictl(args...)
			return nil
		},
	})
}

func initSystemCtr(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use: "ctr",
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		Short: "Containerd CLI",
		Long: `Ctr is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			k3s.ExecCtr(os.Stdout, os.Stderr, args...)
		},
	})
}

func initSystemKubectl(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "kubectl",
		Short: "Kubernetes CLI",
		Long: `Kubectl is a low-level Kubernetes command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			k3s.ExecKubectl(args...)
			return nil
		},
	})
}

func initSystemHelm(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "helm",
		Short: "Helm CLI",
		Long: `helm is a low-level Kubernetes command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			helm.ExecHelm(args...)
			return nil
		},
	})
}
