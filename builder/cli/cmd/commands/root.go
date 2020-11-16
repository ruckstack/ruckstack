package commands

import (
	"github.com/pkg/profile"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "ruckstack",
	Short:            "Ruckstack CLI",
	Long:             "Ruckstack CLI",
	Version:          "0.9.0",
	TraverseChildren: true,
}

var (
	verboseMode     bool
	profileCpu      bool
	profileMemory   bool
	launchVersion   string
	launchImage     string
	launchUser      string
	launchForcePull bool
)

func init() {
	cobra.OnInitialize(initConfig)

	//document and/or don't fail on arguments handled by the launcher
	RootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	RootCmd.Flags().BoolVar(&profileCpu, "profile-cpu", false, "Track cpu usage")
	RootCmd.Flags().BoolVar(&profileMemory, "profile-memory", false, "Track memory usage")
	RootCmd.Flags().StringVar(&launchVersion, "launch-version", "latest", "Specify the version of the Ruckstack CLI to launch")
	RootCmd.Flags().StringVar(&launchImage, "launch-image", "ruckstack", "Specify the Ruckstack CLI image to launch")
	RootCmd.Flags().BoolVar(&launchForcePull, "launch-force-pull", false, "Force the Ruckstack CLI to re-download the image to launch")
}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}

	if !environment.IsRunningLauncher() {
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "--launch-") {
				ui.Printf("WARNING: %s is only used when running the Ruckstack launcher. It is ignored when running the container directly", arg)
			}
		}
	}
}

func Execute(args []string) error {
	if profileCpu {
		defer profile.Start(profile.ProfilePath(environment.ProjectDir)).Stop()
	}
	if profileMemory {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(environment.ProjectDir)).Stop()
	}

	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}
