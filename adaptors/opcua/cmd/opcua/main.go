package main

import (
	"os"

	"github.com/rancher/octopus/adaptors/opcua/pkg/opcua"
	"github.com/rancher/octopus/pkg/util/version/verflag"
	"github.com/spf13/cobra"
)

const (
	name        = "opcua"
	description = ``
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			return opcua.Run()
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
