package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/license"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
)

func init() {
	var parentCommand = &cobra.Command{
		Use:   "license",
		Short: "Commands for managing your Ruckstack Pro license",
	}

	initLicenseSet(parentCommand)
	initLicenseRemove(parentCommand)

	RootCmd.AddCommand(parentCommand)
}

func initLicenseSet(parent *cobra.Command) {

	var licenseText string

	var cmd = &cobra.Command{
		Use:   "set",
		Short: "Sets the active license",

		RunE: func(cmd *cobra.Command, args []string) error {
			if licenseText == "" {
				licenseText = ui.PromptForString("Enter your license key. If you do not have one, visit https://ruckstack.com/pro", "", ui.NotEmptyCheck)
				fmt.Println()
				fmt.Println()
			}

			return license.SetLicense(licenseText)
		},
	}

	cmd.Flags().StringVar(&licenseText, "license", "", "License key. If not set, you will be prompted.")

	parent.AddCommand(cmd)
}

func initLicenseRemove(parent *cobra.Command) {

	var cmd = &cobra.Command{
		Use:   "remove",
		Short: "Removes the active license",

		RunE: func(cmd *cobra.Command, args []string) error {
			return license.RemoveLicense()
		},
	}

	parent.AddCommand(cmd)
}
