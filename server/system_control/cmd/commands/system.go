package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/helm"
	k3s2 "github.com/ruckstack/ruckstack/server/system_control/internal/k3s"
	"github.com/spf13/cobra"
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
			RequiresRoot: "true",
		},
		Short: "CRI CLI",
		Long: `Crictl is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return k3s2.ExecCrictl(environment.ServerHome, args...)
		},
	})
}

func initSystemCtr(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use: "ctr",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		Short: "Containerd CLI",
		Long: `Ctr is a low-level Containerd command.
NOTE: normally this is not a command that should be run, but can be a useful escape hatch.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return k3s2.ExecCtr(environment.ServerHome, args...)
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
			k3s2.ExecKubectl(args...)
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
			return helm.ExecHelm(args...)
		},
	})
}
