package commands

import (
	"github.com/pkg/profile"
	"github.com/ruckstack/ruckstack/builder/cli/internal/analytics"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
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
	Version:          global_util.RuckstackVersion,
	SilenceUsage:     true,
	SilenceErrors:    true,
	TraverseChildren: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		command := cmd.Name()
		parent := cmd.Parent()
		for parent != nil {
			command = parent.Name() + " " + command
			parent = parent.Parent()
		}

		analytics.TrackCommand(command)

		return nil
	},
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
	RootCmd.Flags().StringVar(&launchVersion, "launch-version", "v"+global_util.RuckstackVersion, "Specify the version of the Ruckstack CLI to launch")
	RootCmd.Flags().StringVar(&launchImage, "launch-image", "ghcr.io/ruckstack/ruckstack", "Specify the Ruckstack CLI image to launch")
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

	askAnalytics := true
	for _, flag := range args {
		if flag == "--help" {
			askAnalytics = false
		}
	}

	foundCommand, _, err := RootCmd.Find(args)
	if err != nil || foundCommand.Parent() == nil {
		askAnalytics = false
	}

	if askAnalytics {
		analytics.Ask()
	}

	err = RootCmd.Execute()

	if err != nil {
		analytics.TrackError(err)
	}

	analytics.WaitGroup.Wait()

	return err
}
