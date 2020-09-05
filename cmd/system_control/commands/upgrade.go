package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/upgrade"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"github.com/spf13/cobra"
)

func init() {
	var file string
	packageConfig, err := util.GetPackageConfig()
	if err != nil {
		fmt.Printf("Error loading package config: %s", err)
		return
	}

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades " + packageConfig.Name,
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return upgrade.Upgrade(file)
		},
	}

	upgradeCmd.Flags().StringVar(&file, "file", "", "Path to upgrade file (required)")
	upgradeCmd.MarkFlagFilename("file")
	upgradeCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(upgradeCmd)
}
