package commands

import (
	"github.com/pkg/profile"
	"github.com/ruckstack/ruckstack/builder/internal/analytics"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/spf13/cobra"
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
	verboseMode   bool
	profileCpu    bool
	profileMemory bool
)

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	RootCmd.Flags().BoolVar(&profileCpu, "profile-cpu", false, "Track cpu usage")
	RootCmd.Flags().BoolVar(&profileMemory, "profile-memory", false, "Track memory usage")
}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
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
