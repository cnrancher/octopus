package physical

import (
	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/linux"
)

// newCentral creates a Bluetooth central device.
func newCentral(name string, options ...ble.Option) (ble.Device, error) {
	return linux.NewDeviceWithName(name, append(options, ble.OptCentralRole())...)
}
