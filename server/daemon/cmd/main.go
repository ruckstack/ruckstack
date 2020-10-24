package main

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/daemon/cmd/commands"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"os"
)

func main() {
	if err := os.Setenv("PATH", os.Getenv("PATH")+":"+environment.ServerHome+"/lib"); err != nil {
		ui.Fatalf("Cannot set path")
	}

	err := commands.Execute(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
