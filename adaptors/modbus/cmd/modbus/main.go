package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/rancher/octopus/adaptors/modbus/pkg/modbus"
	"github.com/rancher/octopus/pkg/util/version/verflag"
)

const (
	name        = "modbus"
	description = ``
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			return modbus.Run()
		},
	}
	verflag.AddFlags(c.Flags())
	return c
}

func main() {
	var c = newCommand()
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
