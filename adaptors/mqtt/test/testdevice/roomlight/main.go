package main

import (
	"os"

	"github.com/rancher/octopus/adaptors/mqtt/test/testdevice/roomlight/cmd"
)

func main() {

	c := cmd.NewCommand()

	if c.Execute() != nil {
		os.Exit(1)
	}
}
