package physical

import (
	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

// ModbusDeviceLimbSyncer is used to sync modebus device to limb.
type ModbusDeviceLimbSyncer func(in *v1alpha1.ModbusDevice) error
