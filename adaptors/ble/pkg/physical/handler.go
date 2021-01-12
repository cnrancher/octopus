package physical

import (
	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
)

// BluetoothDeviceLimSyncer is used to sync ble device to limb.
type BluetoothDeviceLimSyncer func(in *v1alpha1.BluetoothDevice, internalError error) error
