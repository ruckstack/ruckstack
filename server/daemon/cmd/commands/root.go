package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	util "github.com/ruckstack/ruckstack/server/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var verboseMode bool

var rootCmd = &cobra.Command{
	TraverseChildren: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	executable, err := os.Executable()
	util.ExpectNoError(err)
	executable = filepath.Base(executable)

	rootCmd.Use = executable

	packageConfig := environment.PackageConfig
	rootCmd.Short = packageConfig.Name + " Server"
	rootCmd.Version = packageConfig.Version

	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
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
