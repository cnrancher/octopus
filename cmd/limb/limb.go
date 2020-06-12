package limb

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/cmd/decorator"
	"github.com/rancher/octopus/cmd/limb/options"
	"github.com/rancher/octopus/pkg/limb"
	"github.com/rancher/octopus/pkg/util/log/logflag"
	"github.com/rancher/octopus/pkg/util/version/verflag"
)

const (
	name        = "limb"
	description = ``
)

func NewCommand() *cobra.Command {
	var opts = options.NewOptions()

	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			logflag.SetLogger(ctrl.SetLogger)

			return limb.Run(name, opts)
		},
	}

	var nfs = opts.Flags(name)
	verflag.AddFlags(nfs.FlagSet("global"))
	logflag.AddFlags(nfs.FlagSet("global"))

	return decorator.Wrap(c, nfs)
}
