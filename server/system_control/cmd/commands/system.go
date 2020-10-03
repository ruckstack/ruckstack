package commands

import (
	k3s2 "github.com/ruckstack/ruckstack/server/internal/k3s"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/helm"
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
			REQUIRES_ROOT: "true",
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
			REQUIRES_ROOT: "true",
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
			return k3s2.ExecKubectl(environment.ServerHome, args...)
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
