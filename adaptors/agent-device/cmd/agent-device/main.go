package main

import (
	"os"

	agentdevice "github.com/rancher/octopus/adaptors/agent-device/pkg/agent-device"
	"github.com/rancher/octopus/pkg/util/version/verflag"
	"github.com/spf13/cobra"
)

const (
	name        = "agent-device"
	description = ``
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			return agentdevice.Run()
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
