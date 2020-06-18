package ble

import (
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/adaptors/ble/pkg/adaptor"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/util/critical"
)

const (
	Name     = "adaptors.edge.cattle.io/ble"
	Version  = "v1alpha1"
	Endpoint = "ble.sock"
)

// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=bluetoothdevices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=bluetoothdevices/status,verbs=get;update;patch

func Run() error {
	log.Info("Starting")

	var stop = ctrl.SetupSignalHandler()
	var ctx = critical.Context(stop)
	eg, ctx := errgroup.WithContext(ctx)
	stop = ctx.Done()
	svc := adaptor.NewService()
	defer svc.Close()
	eg.Go(func() error {
		// start adaptor to receive requests from Limb
		return connection.Serve(Endpoint, svc, stop)
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
