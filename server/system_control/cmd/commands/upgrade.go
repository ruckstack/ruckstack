package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/upgrade"
	"github.com/spf13/cobra"
)

func init() {
	var file string
	packageConfig := environment.PackageConfig

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades " + packageConfig.Name,
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return upgrade.Upgrade(file)
		},
	}

	upgradeCmd.Flags().StringVar(&file, "file", "", "Path to upgrade file (required)")

	ui.MarkFlagsRequired(upgradeCmd, "file")
	ui.MarkFlagsFilename(upgradeCmd, "file")

	rootCmd.AddCommand(upgradeCmd)
}
