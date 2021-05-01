package commands

import (
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/ops"
	"github.com/spf13/cobra"
)

func init() {
	if environment.PackageConfig.LicenseLevel > 0 {

		var opsCmd = &cobra.Command{
			Use:   "ops",
			Short: "Configures the /ops site",
		}

		initOpsUsersList(opsCmd)
		initOpsUserAdd(opsCmd)
		initOpsUserDelete(opsCmd)

		rootCmd.AddCommand(opsCmd)
	}
}

func initOpsUsersList(parent *cobra.Command) {

	var cmd = &cobra.Command{
		Use:   "list-users",
		Short: "Lists users that can access /ops",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ops.ListUsers()
		},
	}

	parent.AddCommand(cmd)
}

func initOpsUserAdd(parent *cobra.Command) {

	var cmd = &cobra.Command{
		Use:   "add-user",
		Short: "Add user to /ops",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ops.AddUser()
		},
	}

	parent.AddCommand(cmd)
}

func initOpsUserDelete(parent *cobra.Command) {

	var cmd = &cobra.Command{
		Use:   "delete-user",
		Short: "Removes user from /ops",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ops.DeleteUser()
		},
	}

	parent.AddCommand(cmd)
}
