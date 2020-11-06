package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	RequiresRoot = "REQUIRES_ROOT_USER"
)

var verboseMode bool

var rootCmd = &cobra.Command{
	TraverseChildren: true,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Annotations[RequiresRoot] == "true" {
			if !environment.IsRunningAsRoot {
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

	packageConfig := environment.PackageConfig
	rootCmd.Short = packageConfig.Name + " System Control"
	rootCmd.Version = packageConfig.Version

	var serverHome string

	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	rootCmd.Flags().StringVar(&serverHome, "server-home", "", "Enable more detailed output")

}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
