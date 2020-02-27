package dummy

import (
	"context"
	"path/filepath"

	"github.com/rancher/octopus/adaptors/dummy/pkg/adaptor"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/dialer"
	"github.com/rancher/octopus/pkg/adaptor/listener"
)

func Run() error {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var stopChan = make(chan error)

	// start server
	var sc, sl, err = listener.Listen(ctx, filepath.Join(api.AdaptorPath, adaptor.Endpoint))
	if err != nil {
		return err
	}
	go func() {
		var a = adaptor.NewAdaptor(sc)
		stopChan <- a.Serve(sl)
	}()

	// register
	cc, err := dialer.Dial(ctx, api.LimbSocket)
	if err != nil {
		return err
	}
	if err = adaptor.Register(ctx, cc); err != nil {
		return err
	}

	return <-stopChan
}
