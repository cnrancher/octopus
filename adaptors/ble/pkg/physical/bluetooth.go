package physical

import (
	"context"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/pkg/errors"
)

var (
	ErrGATTNotFound                = errors.New("GATT is not found")
	ErrBluetoothPeripheralNotFound = errors.New("Bluetooth peripheral is not found")
)

var (
	central ble.Device
	ctx     context.Context
	cancel  context.CancelFunc
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

// EstablishGATT establishes GATT.
func EstablishGATT() error {
	var dev, err = newCentral("Octopus")
	if err != nil {
		return errors.Wrap(err, "failed to start GATT central")
	}
	central = dev
	return nil
}

// CloseGATT closes GATT.
func CloseGATT() error {
	cancel()
	if central != nil {
		var err = central.Stop()
		if err != nil {
			return errors.Wrap(err, "failed to stop GATT central")
		}
	}
	return nil
}

// ScanBluetoothPeripherals scans Bluetooth peripherals.
func ScanBluetoothPeripherals(endpoint string, timeout time.Duration) ([]*BluetoothPeripheral, error) {
	if central == nil {
		return nil, ErrGATTNotFound
	}

	var peripherals []*BluetoothPeripheral

	if timeout == 0 {
		timeout = 15 * time.Second
	}
	var timeoutCtx, cancelTimeoutCtx = context.WithTimeout(ctx, timeout)
	defer cancelTimeoutCtx()
	var err = central.Scan(timeoutCtx, false, func(adv ble.Advertisement) {
		if endpoint == "" {
			peripherals = append(peripherals, NewBluetoothPeripheral(adv))
		} else if adv.LocalName() == endpoint ||
			adv.Addr().String() == endpoint {
			cancelTimeoutCtx() // stops scanning
			peripherals = append(peripherals, NewBluetoothPeripheral(adv))
		}
	})
	if err != nil && err != context.Canceled {
		if endpoint == "" {
			return nil, errors.Wrap(err, "failed to scan GATT peripherals")
		}
		return nil, errors.Wrapf(err, "failed to scan GATT peripheral %s", endpoint)
	}

	if endpoint != "" && len(peripherals) == 0 {
		return nil, ErrBluetoothPeripheralNotFound
	}

	return peripherals, nil
}
