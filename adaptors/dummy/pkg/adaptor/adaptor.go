package adaptor

import (
	"io"
	"net"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/dummy/pkg/device"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

func NewAdaptor(server *grpc.Server) *Adaptor {
	var a = &Adaptor{
		server: server,
		pool:   device.NewPool(),
	}

	api.RegisterAdaptorServiceServer(a.server, a)
	return a
}

type Adaptor struct {
	server *grpc.Server
	pool   device.Pool
}

func (a *Adaptor) KeepAlive(server api.AdaptorService_KeepAliveServer) error {
	for {
		var _, err = server.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Error(err, "failed to receive keepalive request from Limb")
			return err
		}
	}

	server.SendAndClose(&api.Void{})
	return nil
}

func (a *Adaptor) Connect(server api.AdaptorService_ConnectServer) error {
	for {
		var req, err = server.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Error(err, "failed to receive connect request from Limb")
			return err
		}

		// validate parameters
		var parameters = &device.Parameters{}
		if err := jsoniter.Unmarshal(req.GetParameters(), parameters); err != nil {
			return status.Errorf(codes.Internal, "failed to unmarshal parameters: %v", err)
		}
		if err := parameters.Validate(); err != nil {
			return status.Errorf(codes.InvalidArgument, "failed to validate parameters: %v", err)
		}

		// validate device
		var dummy = &v1alpha1.DummyDevice{}
		if err := jsoniter.Unmarshal(req.GetDevice(), dummy); err != nil {
			return status.Errorf(codes.Internal, "failed to unmarshal device: %v", err)
		}

		if err := a.pool.Apply(dummy, parameters, server); err != nil {
			return status.Errorf(codes.Internal, "failed to apply device: %v", err)
		}
	}

	return nil
}

func (a *Adaptor) Serve(lis net.Listener) error {
	defer func() {
		if a.server != nil {
			a.server.Stop()
		}
	}()

	if a.server != nil {
		return a.server.Serve(lis)
	}
	return nil
}
