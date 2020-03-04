package registration

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/adaptor"
	"github.com/rancher/octopus/pkg/suctioncup/event"
	"github.com/rancher/octopus/pkg/suctioncup/validation"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

var log = ctrl.Log.WithName("suctioncup").WithName("registration")

type Server interface {
	Start(<-chan struct{}) error
}

func NewServer(path string, adaptors adaptor.Adaptors, queue event.Queue) (Server, error) {
	var dir = filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to create socket directory")
	}
	if err := cleanup(dir); err != nil {
		return nil, errors.Wrapf(err, "failed to cleanup socket directory")
	}

	return &server{
		path:               path,
		adaptors:           adaptors,
		adaptorNotifier:    queue.GetAdaptorNotifier(),
		connectionNotifier: queue.GetConnectionNotifier(),
	}, nil
}

type server struct {
	path               string
	adaptors           adaptor.Adaptors
	adaptorNotifier    event.AdaptorNotifier
	connectionNotifier event.ConnectionNotifier
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
	api.RegisterRegistrationServer(srv, s)

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

// implement the Registration rpc protoc
func (s *server) Register(_ context.Context, req *api.RegisterRequest) (*api.Empty, error) {
	log = log.WithValues("adaptor", req.Name)

	defer utilruntime.HandleCrash(handler.NewPanicsLogHandler(log))

	if err := validate(req); err != nil {
		log.Error(err, "rejected the register request")
		return &api.Empty{}, grpcstatus.Error(grpccodes.InvalidArgument, err.Error())
	}

	var adp, err = adaptor.NewAdaptor(api.AdaptorPath, req.Name, req.Endpoint, s.connectionNotifier)
	if err != nil {
		log.Error(err, "unable to connect adaptor")
		return &api.Empty{}, grpcstatus.Errorf(grpcstatus.Code(err), "could not connect the registering adaptor %s", req.Name)
	}
	s.adaptors.Put(adp)

	// use another loop to reduce the blocking of rpc,
	// at the same time, that loop ensures that all links will be updated.
	s.adaptorNotifier.NoticeAdaptorRegistered(adp.GetName())

	return &api.Empty{}, nil
}

func cleanup(socketDir string) error {
	var dir, err = os.Open(socketDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	contents, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, content := range contents {
		if !validation.IsSocketFile(content) {
			continue
		}

		var path = filepath.Join(socketDir, content)
		if stat, err := os.Stat(path); err != nil {
			log.Error(err, "could not stat socket path", "path", path)
			continue
		} else if stat.IsDir() {
			log.Error(nil, "socket path is not a file", "path", path)
			continue
		}

		if err = os.RemoveAll(path); err != nil {
			log.Error(err, "could not remove stale socket file", "path", path)
		}
	}

	return nil
}

func validate(req *api.RegisterRequest) error {
	if !validation.IsSupportedVersion(req.Version) {
		return errors.Errorf("%s is not a supported version", req.Version)
	}
	if !validation.IsQualifiedName(req.Name) {
		return errors.Errorf("the requested name %s is not qualified", req.Name)
	}
	if !validation.IsSocketFile(req.Endpoint) {
		return errors.Errorf("the requested endpoint %s could not be recognized", req.Endpoint)
	}
	return nil
}
