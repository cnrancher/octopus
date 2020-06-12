package options

import (
	cliflag "k8s.io/component-base/cli/flag"
)

type Options struct {
	MetricsAddr          int
	EnableLeaderElection bool
}

func (in *Options) Flags(fsName string) (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet(fsName)
	fs.IntVar(&in.MetricsAddr, "metrics-addr", in.MetricsAddr, "The port is used for serving prometheus metrics")
	fs.BoolVar(&in.EnableLeaderElection, "enable-leader-election", in.EnableLeaderElection, "Enable leader election for controller. Enabling this will ensure there is only one active controller manager.")
	return
}

func NewOptions() *Options {
	return &Options{
		MetricsAddr: 8080,
	}
}
