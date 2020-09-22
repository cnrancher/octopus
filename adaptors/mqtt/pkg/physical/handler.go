package physical

import (
	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
)

// MQTTDeviceLimbSyncer is used to sync mqtt device to limb.
type MQTTDeviceLimbSyncer func(in *v1alpha1.MQTTDevice, internalError error) error
