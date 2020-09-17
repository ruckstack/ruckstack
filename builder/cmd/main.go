package main

import (
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cmd/commands"
	"os"
)

func main() {
	err := commands.Execute(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
