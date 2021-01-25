package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/configmap"
	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "config-data",
		Short: "Access server configuration data",
	}

	initConfigDataList(cmd)
	initConfigDataShow(cmd)

	rootCmd.AddCommand(cmd)

}

func initConfigDataList(parent *cobra.Command) {

	var systemConfigData bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all configuration data",

		RunE: func(cmd *cobra.Command, args []string) error {
			return configmap.List(systemConfigData)
		},
	}

	cmd.Flags().BoolVar(&systemConfigData, "system", false, "List system configurations")

	parent.AddCommand(cmd)
}

func initConfigDataShow(parent *cobra.Command) {

	var systemConfigData bool
	var configName string
	var configKey string

	var cmd = &cobra.Command{
		Use:   "show",
		Short: "Displays configuration data",

		RunE: func(cmd *cobra.Command, args []string) error {
			return configmap.Show(configName, configKey, systemConfigData)
		},
	}

	cmd.Flags().BoolVar(&systemConfigData, "system", false, "Is a system configurations")
	cmd.Flags().StringVar(&configName, "name", "", "Configuration name")
	cmd.Flags().StringVar(&configKey, "key", "", "Configuration key")

	ui.MarkFlagsRequired(cmd, "name", "key")

	parent.AddCommand(cmd)
}
