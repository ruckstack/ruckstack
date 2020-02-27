package main

import (
	"github.com/ruckstack/ruckstack/cmd/installer/cmd"
	"github.com/ruckstack/ruckstack/internal/installer"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"os"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	util.Check(err)
	if currentUser.Username != "root" {
		panic("This installer must be ran as root")
	}

	args := os.Args
	if len(args) == 3 && args[1] == "--upgrade" {
		installPackage := os.Getenv("RUCKSTACK_INSTALL_PACKAGE")
		if installPackage == "" {
			installPackage, err = os.Executable()
			util.Check(err)
		}

		installer.Upgrade(installPackage, args[2])

		os.Exit(0)
	}

	cmd.Execute()
}
