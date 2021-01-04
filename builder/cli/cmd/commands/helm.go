package commands

import (
	"github.com/ruckstack/ruckstack/builder/cli/internal/helm"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var helmCommand = &cobra.Command{
		Use:   "helm",
		Short: "Commands for interacting with the Helm repository",
	}

	initRepoGroup(helmCommand)

	initReIndex(helmCommand)

	RootCmd.AddCommand(helmCommand)
}

func initRepoGroup(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "repo",
		Short: "Commands for interacting with the Helm repository configuration",
	}

	initRepoAdd(cmd)
	initRepoRemove(cmd)

	parent.AddCommand(cmd)

}

func initRepoAdd(parent *cobra.Command) {
	var name string
	var url string
	var username string
	var password string

	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Adds a new repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helm.AddRepository(name, url, username, password); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of repository")
	cmd.Flags().StringVar(&url, "url", "", "URL to repository")
	cmd.Flags().StringVar(&username, "username", "", "Chart repository username")
	cmd.Flags().StringVar(&password, "password", "", "Chart repository password")

	ui.MarkFlagsRequired(cmd, "name", "url")

	parent.AddCommand(cmd)
}

func initRepoRemove(parent *cobra.Command) {
	var name string

	var cmd = &cobra.Command{
		Use:   "remove",
		Short: "Removes a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helm.RemoveRepository(name); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of repository")

	parent.AddCommand(cmd)
}

func initReIndex(parent *cobra.Command) {
	var cmd = &cobra.Command{
		Use:   "re-index",
		Short: "Refreshes list of available helm charts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helm.ReIndex(); err != nil {
				return err
			}

			return nil
		},
	}

	parent.AddCommand(cmd)
}
