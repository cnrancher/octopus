package v1alpha1

const (
	// Healthy means that the adaptor is healthy
	Healthy = "Healthy"

	// Unhealthy means that the adaptor is unhealthy
	Unhealthy = "Unhealthy"
)

const (
	// Version is the current version of the API supported by Limb
	Version = "v1alpha1"

	// AdaptorPath is the folder the Adaptor is expecting sockets to be on
	AdaptorPath = "/var/lib/octopus/adaptors/"

	// LimbSocket is the path of the Limb registry socket
	LimbSocket = AdaptorPath + "limb.socket"
)

var SupportedVersions = []string{Version}
