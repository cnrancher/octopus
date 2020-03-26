package main

import (
	"os"

	"github.com/rancher/octopus/adaptors/modbus/pkg/modbus"
	"github.com/rancher/octopus/pkg/util/version/verflag"
	"github.com/spf13/cobra"
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use: "modbus",
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
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
