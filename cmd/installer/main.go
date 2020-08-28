package main

import (
	"fmt"
	"github.com/ruckstack/ruckstack/cmd/installer/commands"
	"github.com/ruckstack/ruckstack/internal/installer"
	"os"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		mainFailed("Error getting user:", err)
	}

	if currentUser.Username != "root" {
		mainFailed("This installer must be ran as root")
	}

	args := os.Args
	if len(args) == 3 && args[1] == "--upgrade" {
		installPackage := os.Getenv("RUCKSTACK_INSTALL_PACKAGE")
		if installPackage == "" {
			installPackage, err = os.Executable()
			if err != nil {
				mainFailed(err)
			}
		}

		installer.Upgrade(installPackage, args[2])

		os.Exit(0)
	}

	err = commands.Execute(os.Args[1:])

	if err != nil {
		mainFailed(err)
	}
}

func mainFailed(messages ...interface{}) {
	fmt.Println(messages)
	os.Exit(-1)
}
