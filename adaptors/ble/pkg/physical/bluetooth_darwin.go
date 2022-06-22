package physical

import (
	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/darwin"
)

// newCentral creates a Bluetooth central device.
func newCentral(name string, options ...ble.Option) (ble.Device, error) {
	return darwin.NewDevice(append(options, ble.OptCentralRole())...)
}
