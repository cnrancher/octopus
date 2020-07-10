package connection

import (
	"k8s.io/apimachinery/pkg/util/runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
)

// Serve provides the connection service `svc` on /var/lib/octopus/adaptors/`endpoint`,
// and is affected by `stop` chan.
func Serve(endpoint string, svc api.ConnectionServer, stop <-chan struct{}) error {
	var srv = NewServer(endpoint, releaseSocket(endpoint, svc))
	return srv.Start(stop)
}

// releaseSocket wraps the ConnectionServer instance as releaseSocketServer.
func releaseSocket(endpoint string, svc api.ConnectionServer) *releaseSocketServer {
	return &releaseSocketServer{
		endpoint: endpoint,
		svc:      svc,
	}
}

// releaseSocketServer decorates the Connect method of ConnectionServer to release the socket file when crashed,
// but panic still continue.
type releaseSocketServer struct {
	endpoint string
	svc      api.ConnectionServer
}

func (s *releaseSocketServer) Connect(server api.Connection_ConnectServer) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(s.endpoint))

	return s.svc.Connect(server)
}
