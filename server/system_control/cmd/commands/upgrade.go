package commands

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/upgrade"
	"github.com/spf13/cobra"
)

func init() {
	var file string
	packageConfig, err := common2.GetPackageConfig()
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
