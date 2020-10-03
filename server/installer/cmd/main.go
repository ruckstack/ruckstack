package main

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/installer/cmd/commands"
	"os"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		ui.Fatalf("Cannot determine current user: %s", err)
	}

	if currentUser.Name != "root" {
		ui.Fatalf("This installer must be ran as root")
	}

	if err := commands.Execute(os.Args[1:]); err != nil {
		ui.Fatalf("Error executing %s: %s", os.Args[0], err)
	}
}
