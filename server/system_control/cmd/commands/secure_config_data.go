package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/secrets"
	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "secure-config-data",
		Short: "Access secure server configuration data",
	}

	initSecretsList(cmd)
	initSecretsShow(cmd)

	rootCmd.AddCommand(cmd)

}

func initSecretsList(parent *cobra.Command) {

	var systemSecrets bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all secured data",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return secrets.List(systemSecrets)
		},
	}

	cmd.Flags().BoolVar(&systemSecrets, "system", false, "List system configurations")

	parent.AddCommand(cmd)
}

func initSecretsShow(parent *cobra.Command) {

	var systemSecrets bool
	var secretName string
	var secretKey string

	var cmd = &cobra.Command{
		Use:   "show",
		Short: "Displays secure data",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return secrets.Show(secretName, secretKey, systemSecrets)
		},
	}

	cmd.Flags().BoolVar(&systemSecrets, "system", false, "Is a system configurations")
	cmd.Flags().StringVar(&secretName, "name", "", "Configuration name")
	cmd.Flags().StringVar(&secretKey, "key", "", "Configuration key")

	ui.MarkFlagsRequired(cmd, "name", "key")

	parent.AddCommand(cmd)
}
