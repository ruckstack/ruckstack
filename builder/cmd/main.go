package main

import (
	"github.com/ruckstack/ruckstack/builder/cmd/commands"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"os"
)

func main() {
	err := commands.Execute(os.Args[1:])

	environment.Cleanup()
	if err != nil {
		ui.Fatalf("Error executing %s: %s", os.Args[0], err)
	}
}
