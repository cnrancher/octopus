package dialer

import (
	"context"
	"net"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func Dial(ctx context.Context, socketPath string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var dialerOptions = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (conn net.Conn, err error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	}

	var conn, err = grpc.DialContext(ctx, socketPath, dialerOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial adaptor: %s", socketPath)
	}

	return conn, nil
}
