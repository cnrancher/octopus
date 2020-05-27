package connection

import (
	"os"
	"path/filepath"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
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
	defer func() {
		if r := recover(); r != nil {
			var socketPath = filepath.Join(api.AdaptorPath, s.endpoint)
			if pi, err := os.Stat(socketPath); err == nil && !pi.IsDir() && pi.Mode()&os.ModeSocket != 0 {
				_ = os.RemoveAll(socketPath)
			}

			// NB(thxCode) the purpose in this recover is to clean up the remaining socket file,
			// which ensure the adaptor can be restarted again. For irreversible panic,
			// returning an error is unable to solve the internal problem, and may lead to other serious errors.
			// So just panic to upper level.
			panic(r)
		}
	}()
	return s.svc.Connect(server)
}
