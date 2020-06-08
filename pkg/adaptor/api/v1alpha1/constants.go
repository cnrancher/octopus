package v1alpha1

const (
	// Version is the current version of the API supported by Limb
	Version = "v1alpha1"

	// AdaptorPath is the folder the adaptor is expecting sockets to be on
	AdaptorPath = "/var/lib/octopus/adaptors/"

	// SocketSuffix is the suffix of the socket
	SocketSuffix = ".sock"

	// LimbSocket is the path of the Limb registry socket
	LimbSocket = AdaptorPath + "limb" + SocketSuffix
)

var SupportedVersions = []string{Version}
