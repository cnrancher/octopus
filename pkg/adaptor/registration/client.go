package registration

import (
	"context"
	"net"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

func Register(ctx context.Context, request api.RegisterRequest) error {
	var cliOptions = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (conn net.Conn, err error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	}

	var setupCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var conn, err = grpc.DialContext(setupCtx, api.LimbSocket, cliOptions...)
	if err != nil {
		return errors.Wrapf(err, "failed to dial Limb %s", api.LimbSocket)
	}

	// register adaptor
	if _, err := api.NewRegistrationClient(conn).Register(ctx, &request); err != nil {
		return errors.Wrapf(err, "failed to register to Limb")
	}

	sockWatcher, err := newSocketWatcher()
	if err != nil {
		return errors.Wrapf(err, "failed to create socket watcher")
	}

	// watch limb socket
	return sockWatcher.Watch(ctx.Done())
}
