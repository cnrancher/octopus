package physical

import (
	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
)

// DummySpecialDeviceLimbSyncer is used to sync physical special device.
type DummySpecialDeviceLimbSyncer func(in *v1alpha1.DummySpecialDevice) error

// DummyProtocolDeviceLimbSyncer is used to sync physical special device.
type DummyProtocolDeviceLimbSyncer func(in *v1alpha1.DummyProtocolDevice) error
