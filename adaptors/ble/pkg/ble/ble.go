package ble

import (
	"github.com/rancher/octopus/adaptors/ble/pkg/adaptor"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/util/critical"
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Name     = "adaptors.edge.cattle.io/ble"
	Version  = "v1alpha1"
	Endpoint = "ble.sock"
)

// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=bluetoothdevices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=bluetoothdevices/status,verbs=get;update;patch

func Run() error {
	var stop = ctrl.SetupSignalHandler()
	var ctx = critical.Context(stop)
	eg, ctx := errgroup.WithContext(ctx)
	stop = ctx.Done()
	eg.Go(func() error {
		// start adaptor to receive requests from Limb
		return connection.Serve(Endpoint, adaptor.NewService(), stop)
	})
	eg.Go(func() error {
		// register adaptor to Limb
		return registration.Register(ctx, api.RegisterRequest{
			Name:     Name,
			Version:  Version,
			Endpoint: Endpoint,
		})
	})
	return eg.Wait()
}
