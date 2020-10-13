package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/installer/internal/install_file"
	"github.com/spf13/cobra"
	"os"
	"os/user"
)

var (
	verboseMode        bool
	installPackagePath string
	extractOnly        bool

	installOptions install_file.InstallOptions
	installFile    *install_file.InstallFile
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Installs application",

	RunE: func(cmd *cobra.Command, args []string) error {
		if extractOnly {
			ui.Println("EXTRACT ONLY SELECTED")
			ui.Println("------------------------------------")

			if installOptions.TargetDir == "" {
				ui.Fatal("--install-path is required for extract-only")
			}
			if installOptions.TargetDir == "" {
				ui.Fatal("--admin-group is required for extract-only")
			}

			localConfig := &config.LocalConfig{
				AdminGroup:  installOptions.AdminGroup,
				BindAddress: "127.0.0.1",
			}

			return installFile.Extract(installOptions.TargetDir, localConfig)
		}

		currentUser, err := user.Current()
		if err != nil {
			ui.Fatalf("Cannot determine current user: %s", err)
		}

		if currentUser.Name != "root" {
			ui.Fatalf("This installer must be ran as root")
		}

		if installOptions.TargetDir != "" {
			_, err := config.LoadPackageConfig(installOptions.TargetDir)
			if err == nil {
				ui.VPrintf("%s is an existing install. Upgrading...", installOptions.TargetDir)
				return installFile.Upgrade(installOptions)
			} else {
				ui.Fatalf("Error checking path %s: %s", installOptions.TargetDir, err)
			}
		}

		return installFile.Install(installOptions)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolVar(&verboseMode, "verbose", false, "Enable more detailed output")
	rootCmd.Flags().StringVar(&installOptions.TargetDir, "install-path", "", "Install path")
	rootCmd.Flags().StringVar(&installOptions.AdminGroup, "admin-group", "", "Administrator group")
	rootCmd.Flags().StringVar(&installOptions.BindAddress, "bind-address", "", "IP address to bind to")
	rootCmd.Flags().StringVar(&installOptions.JoinToken, "join-token", "", "Token for joining cluster")

	rootCmd.Flags().BoolVar(&extractOnly, "extract-only", false, "INTERNAL: only extract the files, don't install")
	rootCmd.Flag("extract-only").Hidden = true
}

func initConfig() {
	if verboseMode {
		ui.SetVerbose(true)
	}
}

func Execute(args []string) error {
	installPackagePath = os.Getenv("RUCKSTACK_INSTALL_PACKAGE")
	if installPackagePath == "" {
		installPackagePath = os.Args[0]
	}

	var err error
	installFile, err = install_file.Parse(installPackagePath)
	if err != nil {
		return err
	}

	rootCmd.Short = fmt.Sprintf("Installs %s %s", installFile.PackageConfig.Name, installFile.PackageConfig.Version)
	rootCmd.Version = installFile.PackageConfig.Version

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
