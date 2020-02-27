package options

import (
	cliflag "k8s.io/component-base/cli/flag"
)

type Options struct {
	MetricsAddr int
	NodeName    string
}

func (in *Options) Flags(fsName string) (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet(fsName)
	fs.IntVar(&in.MetricsAddr, "metrics-addr", in.MetricsAddr, "The port is used for serving prometheus metrics")
	fs.StringVar(&in.NodeName, "node-name", in.NodeName, "The name of the node, using 'NODE_NAME' environment variable is the same")
	return
}

func NewOptions() *Options {
	return &Options{
		MetricsAddr: 8080,
	}
}
