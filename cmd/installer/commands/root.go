package commands

import (
	"archive/zip"
	"fmt"
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/installer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
)

var packageConfig *internal.PackageConfig
var zipReader *zip.ReadCloser
var installerArgs *installer.InstallerArgs

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Installs application",

	RunE: func(cmd *cobra.Command, args []string) error {
		installer.Install(packageConfig, installerArgs, zipReader)

		return nil
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

	rootCmd.Flags().StringVar(&installerArgs.InstallPath, "install-path", "", "Install path")
	rootCmd.Flags().StringVar(&installerArgs.AdminGroup, "admin-group", "", "Administrator group")
	rootCmd.Flags().StringVar(&installerArgs.BindAddress, "bind-address", "", "IP address to bind to")
}

func Execute(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
