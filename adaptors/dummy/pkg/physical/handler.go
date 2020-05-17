package physical

import (
	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
)

// SpecialDeviceSyncer is used to sync physical special device.
type SpecialDeviceSyncer func(in *v1alpha1.DummySpecialDevice)

// ProtocolDeviceSyncer is used to sync physical special device.
type ProtocolDeviceSyncer func(in *v1alpha1.DummyProtocolDevice)
