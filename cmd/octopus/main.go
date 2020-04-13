package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rancher/octopus/cmd/brain"
	"github.com/rancher/octopus/cmd/limb"
	_ "github.com/rancher/octopus/pkg/util/log/handler"
	"github.com/rancher/octopus/pkg/util/version/verflag"
)

const (
	name        = "octopus"
	description = ``
)

var allCommands = []*cobra.Command{
	brain.NewCommand(),
	limb.NewCommand(),
}

func main() {
	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)

			var (
				basename  = filepath.Base(os.Args[0])
				targetCmd *cobra.Command
			)
			for _, cmd := range allCommands {
				if cmd.Name() == basename {
					targetCmd = cmd
					break
				}
				for _, alias := range cmd.Aliases {
					if alias == basename {
						targetCmd = cmd
						break
					}
				}
			}
			if targetCmd != nil {
				return targetCmd.Execute()
			}
			return cmd.Help()
		},
	}
	c.AddCommand(allCommands...)
	verflag.AddFlags(c.Flags())

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
