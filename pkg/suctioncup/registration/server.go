package registration

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
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

	var socketWatcher, err = newSocketWatcher(log.WithName("watcher"), dir, adaptors, queue.GetAdaptorNotifier())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create socket watcher")
	}

	return &server{
		path:         path,
		connNotifier: queue.GetConnectionNotifier(),
		sockWatcher:  socketWatcher,
	}, nil
}

type server struct {
	path         string
	connNotifier event.ConnectionNotifier
	sockWatcher  *socketWatcher
}

func (s *server) Start(stop <-chan struct{}) error {
	var lis, err = net.Listen("unix", s.path)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on: %s", s.path)
	}
	defer func() {
		_ = lis.Close()
	}()

	// start adaptor socket watcher
	go s.sockWatcher.Start(stop)

	// start grpc server
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
	var log = log.WithValues("adaptor", req.Name)

	defer utilruntime.HandleCrash(handler.NewPanicsLogHandler(log))

	if err := validate(req); err != nil {
		log.Error(err, "Rejected the register request")
		return &api.Empty{}, grpcstatus.Error(grpccodes.InvalidArgument, err.Error())
	}

	var adp, err = adaptor.NewAdaptor(api.AdaptorPath, req.Name, req.Endpoint, s.connNotifier)
	if err != nil {
		log.Error(err, "Unable to connect adaptor")
		return &api.Empty{}, grpcstatus.Errorf(grpcstatus.Code(err), "could not connect the registering adaptor %s", req.Name)
	}
	if err := s.sockWatcher.Watch(adp); err != nil {
		log.Error(err, "Unable to watch adaptor's socket")
		return &api.Empty{}, grpcstatus.Errorf(grpccodes.Internal, "could not watch the socket of registering adaptor %s", req.Name)
	}

	return &api.Empty{}, nil
}

func cleanup(socketDir string) error {
	var log = log.WithName("cleaner")

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
		var path = filepath.Join(socketDir, content)
		var pi, err = os.Stat(path)
		if err != nil {
			log.Error(err, "Cannot get stat of path", "path", path)
			continue
		}
		if pi.IsDir() {
			log.V(0).Info("Ignore dir", "path", path)
			continue
		}
		if strings.HasPrefix(pi.Name(), ".") {
			log.V(0).Info("Ignore ignoring file", "path", path)
			continue
		}
		if pi.Mode()&os.ModeSocket == 0 {
			log.V(0).Info("Ignore non socket file", "path", path)
			continue
		}
		err = os.RemoveAll(path)
		if err != nil {
			log.Error(err, "Cannot remove stale socket file", "path", path)
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
	return nil
}
