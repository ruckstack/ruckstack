package main

import (
	"github.com/ruckstack/ruckstack/builder/cli/cmd/commands"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
)

func main() {
	err := commands.Execute(os.Args[1:])

	if err != nil {
		ui.Fatalf("Error executing %s: %s", os.Args[0], err)
	}
}
