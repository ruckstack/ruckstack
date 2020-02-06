package cmd

import (
	"github.com/ruckstack/ruckstack/internal/system-control/upgrade"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use: "upgrade",
	Annotations: map[string]string{
		ANNOTATION_REQUIRES_ROOT: "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		upgrade.Upgrade(file)
	},
}

var file string

func init() {
	packageConfig := util.GetPackageConfig()
	upgradeCmd.Short = "Upgrades " + packageConfig.Name

	upgradeCmd.Flags().StringVar(&file, "file", "", "Path to upgrade file (required)")
	util.Check(upgradeCmd.MarkFlagFilename("file"))
	util.Check(upgradeCmd.MarkFlagRequired("file"))

	rootCmd.AddCommand(upgradeCmd)
}
