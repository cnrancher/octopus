package physical

import api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, configuration interface{}) error
}
