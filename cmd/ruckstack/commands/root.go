package commands

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/ui"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "ruckstack",
	Short:            "Ruckstack CLI",
	Long:             `Ruckstack CLI`,
	Version:          "0.8.3",
	TraverseChildren: true,
}

var verboseMode bool
var useVersion string
var imageName string

func init() {
	cobra.OnInitialize(initConfig)

	//document and/or don't fail on arguments handled by the launcher
	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	rootCmd.Flags().StringVar(&useVersion, "launch-version", "latest", "Specify the version of the Ruckstack cli to use")
	rootCmd.Flags().StringVar(&imageName, "launch-image", "ruckstack", "Specify the Ruckstack cli image to use")
}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}

	if !util.IsRunningLauncher() {
		if useVersion != "latest" {
			ui.Println("WARNING: --launch-version is only used when running the Ruckstack launcher. It is ignored when running the container directly")
		}

		if imageName != "ruckstack" {
			ui.Println("WARNING: --launch-image is only used when running the Ruckstack launcher. It is ignored when running the container directly")
		}
	}
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
