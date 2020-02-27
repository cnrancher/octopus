package listener

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

func Listen(ctx context.Context, socketPath string) (*grpc.Server, net.Listener, error) {
	var socket, err = net.Listen("unix", socketPath)
	if err != nil {
		return nil, nil, err
	}
	var serverOptions = []grpc.ServerOption{
		grpc.ConnectionTimeout(10 * time.Second),
	}
	return grpc.NewServer(serverOptions...), socket, nil
}
