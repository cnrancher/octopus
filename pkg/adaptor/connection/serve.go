package connection

import (
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

// Serve provides the connection service `svc` on /var/lib/octopus/adaptors/`endpoint`,
// and is affected by `stop` chan
func Serve(endpoint string, svc api.ConnectionServer, stop <-chan struct{}) error {
	var srv = NewServer(endpoint, svc)
	return srv.Start(stop)
}
