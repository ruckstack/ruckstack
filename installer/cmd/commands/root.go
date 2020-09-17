package commands

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/common"
	internal2 "github.com/ruckstack/ruckstack/installer/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
)

var packageConfig *common.PackageConfig
var zipReader *zip.ReadCloser
var installerArgs *internal2.InstallerArgs

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Installs application",

	RunE: func(cmd *cobra.Command, args []string) error {
		return internal2.Install(packageConfig, installerArgs, zipReader)
	},
}

func init() {
	if len(os.Args) == 3 && os.Args[1] == "--upgrade" {
		return
	}

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
			packageConfig = &common.PackageConfig{}
			err = decoder.Decode(packageConfig)
			if err != nil {
				panic(err)
			}
		}
	}

	installerArgs = new(internal2.InstallerArgs)

	rootCmd.Short = fmt.Sprintf("Installs %s %s", packageConfig.Name, packageConfig.Version)
	rootCmd.Version = packageConfig.Version

	rootCmd.Flags().StringVar(&installerArgs.InstallPath, "install-path", "", "Install path")
	rootCmd.Flags().StringVar(&installerArgs.AdminGroup, "admin-group", "", "Administrator group")
	rootCmd.Flags().StringVar(&installerArgs.BindAddress, "bind-address", "", "IP address to bind to")
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
