package physical

import (
	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
)

// OPCUADeviceLimbSyncer is used to sync opcua device to limb.
type OPCUADeviceLimbSyncer func(in *v1alpha1.OPCUADevice) error
