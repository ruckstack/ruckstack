package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	REQUIRES_ROOT = "REQUIRES_ROOT_USER"
)

var rootCmd = &cobra.Command{
	TraverseChildren: true,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Annotations[REQUIRES_ROOT] == "true" {
			if environment.IsRunningAsRoot {
				return fmt.Errorf("command %s must be run as sudo or root", cmd.Name())
			}

		}
		return nil
	},
}

func init() {
	executable, err := os.Executable()
	util.ExpectNoError(err)
	executable = filepath.Base(executable)

	packageConfig := environment.PackageConfig

	rootCmd.Use = executable
	rootCmd.Short = packageConfig.Name + " System Control"
	rootCmd.Version = packageConfig.Version
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
