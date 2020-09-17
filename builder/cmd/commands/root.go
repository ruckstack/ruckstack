package commands

import (
	"github.com/ruckstack/ruckstack/builder/internal/util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
	"os"
	"strings"
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
var launchVersion string
var launchImage string
var launchUser string
var launchForcePull bool

func init() {
	cobra.OnInitialize(initConfig)

	//document and/or don't fail on arguments handled by the launcher
	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	rootCmd.Flags().StringVar(&launchVersion, "launch-version", "latest", "Specify the version of the Ruckstack CLI to launch")
	rootCmd.Flags().StringVar(&launchImage, "launch-image", "ruckstack", "Specify the Ruckstack CLI image to launch")
	rootCmd.Flags().StringVar(&launchUser, "launch-user", "", "Specify `user:group` to run the image under")
	rootCmd.Flags().BoolVar(&launchForcePull, "launch-force-pull", false, "Force the Ruckstack CLI to re-download the image to launch")
}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}

	if !util.IsRunningLauncher() {
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "--launch-") {
				ui.Printf("WARNING: %s is only used when running the Ruckstack launcher. It is ignored when running the container directly", arg)
			}
		}
	}
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
