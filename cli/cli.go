package main

import (
	"log"
	"os"

	"github.com/Perkovec/StatiStream/cli/commands"
	"github.com/hashicorp/cli"
)

func main() {
	cmd := &commands.CommandsFactory{}

	c := cli.NewCLI("StatiStream", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"stream": cmd.NewStreamCommand,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
