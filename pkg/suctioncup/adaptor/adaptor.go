package adaptor

import (
	"context"
	"net"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/pkg/suctioncup/connection"
	"github.com/rancher/octopus/pkg/suctioncup/event"
)

type Adaptor interface {
	// GetName returns the name of adaptor
	GetName() string

	// GetEndpoint returns the endpoint of adaptor
	GetEndpoint() string

	// Stop stops the adaptor and deletes all connections
	Stop() error

	// CreateConnection creates a connection by name
	CreateConnection(name types.NamespacedName) (overwritten bool, conn connection.Connection, err error)

	// DeleteConnection deletes the connection of name
	DeleteConnection(name types.NamespacedName) (exist bool)
}

func NewAdaptor(dir, name, endpoint string, notifier event.ConnectionNotifier) (Adaptor, error) {
	var socketPath = filepath.Join(dir, endpoint)

	var cliOptions = []grpc.DialOption{
		// TODO add keepalive supported
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (conn net.Conn, err error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	}

	var setupCtx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var conn, err = grpc.DialContext(setupCtx, socketPath, cliOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial adaptor: %s", socketPath)
	}

	return &adaptor{
		name:       name,
		endpoint:   endpoint,
		clientConn: conn,
		conns:      connection.NewConnections(),
		notifier:   notifier,
	}, nil
}

type adaptor struct {
	name       string
	endpoint   string
	clientConn *grpc.ClientConn
	conns      connection.Connections
	notifier   event.ConnectionNotifier
}

func (a *adaptor) GetName() string {
	return a.name
}

func (a *adaptor) GetEndpoint() string {
	return a.endpoint
}

func (a *adaptor) Stop() error {
	a.conns.Cleanup()

	var err = a.clientConn.Close()
	if status, ok := grpcstatus.FromError(err); ok {
		// NB(thxCode) after multiple shutdowns, it will not cause a leak,
		// which is an informative error.
		if status.Code() == grpccodes.Canceled {
			return nil
		}
	}
	return err
}

func (a *adaptor) CreateConnection(name types.NamespacedName) (overwritten bool, conn connection.Connection, err error) {
	conn = a.conns.Get(name)
	if conn != nil {
		if !conn.IsStop() {
			return true, conn, nil
		}
	}
	conn, err = connection.NewConnection(a.name, name, a.clientConn, a.notifier)
	if err != nil {
		return false, nil, err
	}
	return a.conns.Put(conn), conn, nil
}

func (a *adaptor) DeleteConnection(name types.NamespacedName) bool {
	return a.conns.Delete(name)
}
