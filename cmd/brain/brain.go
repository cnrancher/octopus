package brain

import (
	"github.com/spf13/cobra"

	"github.com/rancher/octopus/cmd/brain/options"
	"github.com/rancher/octopus/cmd/decorator"
	"github.com/rancher/octopus/pkg/brain"
	"github.com/rancher/octopus/pkg/util/log/logflag"
	"github.com/rancher/octopus/pkg/util/version/verflag"
)

const (
	name        = "brain"
	description = ``
)

func NewCommand() *cobra.Command {
	opts := options.NewOptions()

	c := &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
			logflag.Configure()

			return brain.Run(name, opts)
		},
	}

	nfs := opts.Flags(name)
	verflag.AddFlags(nfs.FlagSet("global"))
	logflag.AddFlags(nfs.FlagSet("global"))

	return decorator.Wrap(c, nfs)
}
