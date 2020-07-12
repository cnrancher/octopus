package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/rancher/octopus/adaptors/ble/pkg/ble"
	"github.com/rancher/octopus/pkg/adaptor/log"
	_ "github.com/rancher/octopus/pkg/util/log/handler"
	"github.com/rancher/octopus/pkg/util/log/logflag"
	"github.com/rancher/octopus/pkg/util/version/verflag"
)

const (
	name        = "ble"
	description = "The BLE adaptor defines the device configuration and the attributes of connected BLE device"
)

func newCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			logflag.SetLogger(log.SetLogger)

			return ble.Run()
		},
	}

	verflag.AddFlags(c.Flags())
	logflag.AddFlags(c.Flags())
	return c
}

func main() {
	var c = newCommand()
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
