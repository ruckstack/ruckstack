package commands

import (
	"fmt"
	"github.com/pkg/profile"
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

const (
	RequiresRoot = "REQUIRES_ROOT_USER"
)

var verboseMode bool
var profileCpu bool
var profileMemory bool

var rootCmd = &cobra.Command{
	TraverseChildren: true,
	SilenceUsage:     true,
	SilenceErrors:    true,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Annotations[RequiresRoot] == "true" {
			if !config.IsRunningAsRoot {
				return fmt.Errorf("command %s must be run as sudo or root", cmd.Name())
			}

		}
		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	executable, err := os.Executable()
	util.ExpectNoError(err)
	executable = filepath.Base(executable)

	rootCmd.Use = executable

	packageConfig := config.PackageConfig
	rootCmd.Short = packageConfig.Name + " System Control"
	rootCmd.Version = packageConfig.Version

	var serverHome string

	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	rootCmd.Flags().StringVar(&serverHome, "server-home", "", "Enable more detailed output")

	rootCmd.Flags().BoolVar(&profileCpu, "profile-cpu", false, "Track cpu usage")
	rootCmd.Flags().BoolVar(&profileMemory, "profile-memory", false, "Track memory usage")

}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}
}

func Execute(args []string) error {
	for _, arg := range args {
		if arg == "--profile-cpu" {
			defer profile.Start(profile.ProfilePath(fmt.Sprintf("%s/data/profile/cpu-%d", config.ServerHome, time.Now().Unix()))).Stop()
		}
		if arg == "--profile-memory" {
			defer profile.Start(profile.MemProfile, profile.ProfilePath(fmt.Sprintf("%s/data/profile/memory-%d", config.ServerHome, time.Now().Unix()))).Stop()
		}
	}

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
