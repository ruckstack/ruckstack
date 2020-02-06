package cmd

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/installer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var packageConfig *internal.PackageConfig
var zipReader *zip.ReadCloser
var installerArgs *installer.InstallerArgs

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Installs application",
	//Long:  `Ruckstack CLI`,

	Run: func(cmd *cobra.Command, args []string) {
		installer.Install(packageConfig, installerArgs, zipReader)
	},
}

func init() {
	if len(os.Args) == 3 && os.Args[1] == "--upgrade" {
		return
	}

	cobra.OnInitialize(initConfig)

	installPackage := os.Getenv("RUCKSTACK_INSTALL_PACKAGE")
	if installPackage == "" {
		installPackage = os.Args[0]
	}

	var err error
	zipReader, err = zip.OpenReader(installPackage)
	if err != nil {
		panic(err)
	}

	for _, zipFile := range zipReader.File {
		if zipFile.Name == ".package.config" {
			fileReader, err := zipFile.Open()
			if err != nil {
				panic(err)
			}

			decoder := yaml.NewDecoder(fileReader)
			packageConfig = &internal.PackageConfig{}
			err = decoder.Decode(packageConfig)
			if err != nil {
				panic(err)
			}
		}
	}

	installerArgs = new(installer.InstallerArgs)

	rootCmd.Short = fmt.Sprintf("Installs %s %s", packageConfig.Name, packageConfig.Version)
	rootCmd.Version = packageConfig.Version

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s-installer.yaml)", packageConfig.Id))

	rootCmd.Flags().StringVar(&installerArgs.InstallPath, "install-path", "", "Install path")
	rootCmd.Flags().StringVar(&installerArgs.AdminGroup, "admin-group", "", "Administrator group")
	rootCmd.Flags().StringVar(&installerArgs.BindAddress, "bind-address", "", "IP address to bind to")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".installer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("." + packageConfig.Id + "-installer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
