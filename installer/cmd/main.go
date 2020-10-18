package main

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/installer/cmd/commands"
	"os"
)

func main() {
	if err := commands.Execute(os.Args[1:]); err != nil {
		ui.Fatalf("Error executing %s: %s", os.Args[0], err)
	}
}
