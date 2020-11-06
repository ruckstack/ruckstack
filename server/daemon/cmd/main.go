package main

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/daemon/cmd/commands"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	"os"
	"time"
)

func main() {
	fmt.Printf("Starting main at %d", time.Now().Unix())
	//include lib dir in path
	if err := os.Setenv("PATH", os.Getenv("PATH")+":"+environment.ServerHome+"/lib"); err != nil {
		ui.Fatalf("Cannot configure path: %s", err)
	}

	err := commands.Execute(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
