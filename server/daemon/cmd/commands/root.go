package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	util "github.com/ruckstack/ruckstack/server/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var verboseMode bool

var rootCmd = &cobra.Command{
	TraverseChildren: true,
}

func init() {
	fmt.Printf("root init %d", time.Now().Unix())

	cobra.OnInitialize(initConfig)

	executable, err := os.Executable()
	util.ExpectNoError(err)
	executable = filepath.Base(executable)

	rootCmd.Use = executable

	packageConfig := environment.PackageConfig
	rootCmd.Short = packageConfig.Name + " Server"
	rootCmd.Version = packageConfig.Version

	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")

	//not actually used here. It is looked for specially in environment's init() function. That code runs too early to use this, but we need CLI validation to pass
	var serverHome string
	rootCmd.Flags().StringVar(&serverHome, "server-home", "", "Override the server home directory to use (INTERNAL)")
	rootCmd.Flag("server-home").Hidden = true

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
