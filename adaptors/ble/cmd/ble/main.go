package main

import (
	"os"

	"github.com/rancher/octopus/adaptors/ble/pkg/ble"
	"github.com/rancher/octopus/pkg/util/version/verflag"
	"github.com/spf13/cobra"
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use: "ble",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ble.Run()
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
