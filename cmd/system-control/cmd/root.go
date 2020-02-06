package cmd

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	ANNOTATION_REQUIRES_ROOT = "requiresRoot"
)

var rootCmd = &cobra.Command{
	TraverseChildren: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:/
	//	Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Annotations[ANNOTATION_REQUIRES_ROOT] == "true" {
			currentUser, err := user.Current()
			util.Check(err)
			if currentUser.Username != "root" {
				panic("Requires root")
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	executable, err := os.Executable()
	util.Check(err)
	executable = filepath.Base(executable)

	packageConfig := util.GetPackageConfig()
	rootCmd.Use = executable
	rootCmd.Short = packageConfig.Name + " CLI"
	rootCmd.Version = packageConfig.Version

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/."+executable+".yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)

		// Search config in home directory with name ".system-control" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".server")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
