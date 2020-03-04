package connection

import (
	"net"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

type Server interface {
	Start(<-chan struct{}) error
}

func NewServer(endpoint string, svc api.ConnectionServer) Server {
	return &server{
		path: filepath.Join(api.AdaptorPath, endpoint),
		svc:  svc,
	}
}

type server struct {
	path string
	svc  api.ConnectionServer
}

func (s *server) Start(stop <-chan struct{}) error {
	var lis, err = net.Listen("unix", s.path)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on: %s", s.path)
	}
	defer func() {
		_ = lis.Close()
	}()

	var srvOptions = []grpc.ServerOption{
		grpc.ConnectionTimeout(10 * time.Second),
	}
	var srv = grpc.NewServer(srvOptions...)
	defer srv.Stop()

	// register services
	api.RegisterConnectionServer(srv, s.svc)

	// serve
	var errC = make(chan error)
	go func() {
		errC <- srv.Serve(lis)
	}()

	select {
	case err := <-errC:
		return err
	case <-stop:
		return nil
	}
}
