package modbus

import (
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/adaptors/modbus/pkg/adaptor"
	"github.com/rancher/octopus/adaptors/modbus/pkg/metadata"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/util/critical"
)

// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=modbusdevices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devices.edge.cattle.io,resources=modbusdevices/status,verbs=get;update;patch

func Run() error {
	log.Info("Starting")

	var stop = ctrl.SetupSignalHandler()
	var ctx = critical.Context(stop)
	eg, ctx := errgroup.WithContext(ctx)
	stop = ctx.Done()
	eg.Go(func() error {
		// start adaptor to receive requests from Limb
		return connection.Serve(metadata.Endpoint, adaptor.NewService(), stop)
	})
	eg.Go(func() error {
		// register adaptor to Limb
		return registration.Register(ctx, api.RegisterRequest{
			Name:     metadata.Name,
			Version:  metadata.Version,
			Endpoint: metadata.Endpoint,
		})
	})
	return eg.Wait()
}
