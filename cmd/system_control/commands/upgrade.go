package commands

import (
	"github.com/ruckstack/ruckstack/internal/system_control/upgrade"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"github.com/spf13/cobra"
)

func init() {
	var file string
	packageConfig := util.GetPackageConfig()

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades " + packageConfig.Name,
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			upgrade.Upgrade(file)
			return nil
		},
	}

	upgradeCmd.Flags().StringVar(&file, "file", "", "Path to upgrade file (required)")
	util.Check(upgradeCmd.MarkFlagFilename("file"))
	util.Check(upgradeCmd.MarkFlagRequired("file"))

	rootCmd.AddCommand(upgradeCmd)
}
