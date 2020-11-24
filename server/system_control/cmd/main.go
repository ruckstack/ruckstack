package main

import (
	"fmt"
	"github.com/ruckstack/ruckstack/server/system_control/cmd/commands"
	"os"
)

func main() {
	err := commands.Execute(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
