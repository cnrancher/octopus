package options

import (
	cliflag "k8s.io/component-base/cli/flag"
)

type Options struct {
	AdmissionWebhookAddr    int
	DisableAdmissionWebhook bool
	MetricsAddr             int
	EnableLeaderElection    bool
}

func (in *Options) Flags(fsName string) (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet(fsName)
	fs.IntVar(&in.AdmissionWebhookAddr, "admission-webhook-addr", in.AdmissionWebhookAddr, "The port is used for serving admission server.")
	fs.BoolVar(&in.DisableAdmissionWebhook, "disable-admission-webhook", in.DisableAdmissionWebhook, "Disable admission webhook for controller.")
	fs.IntVar(&in.MetricsAddr, "metrics-addr", in.MetricsAddr, "The port is used for serving prometheus metrics")
	fs.BoolVar(&in.EnableLeaderElection, "enable-leader-election", in.EnableLeaderElection, "Enable leader election for controller. Enabling this will ensure there is only one active controller manager.")
	return
}

func NewOptions() *Options {
	return &Options{
		MetricsAddr: 8080,
	}
}
