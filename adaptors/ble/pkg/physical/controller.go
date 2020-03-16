package physical

import (
	"fmt"
	"strings"
	"time"

	"github.com/bettercap/gatt"
	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct {
	spec   v1alpha1.BluetoothDeviceSpec
	status v1alpha1.BluetoothDeviceStatus
	done   chan struct{}
}

func onStateChanged(d gatt.Device, s gatt.State) {
	logrus.Printf("Bluetooth state: %s", s)
	switch s {
	case gatt.StatePoweredOn:
		logrus.Info("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (c *Controller) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	name := c.spec.Name
	addr := c.spec.MacAddress
	if name != "" && a.LocalName != name {
		return
	}

	if addr != "" && strings.ToUpper(p.ID()) != strings.ToUpper(addr) {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	logrus.Info("Stop scanning and found device ", a.LocalName)
	p.Device().StopScanning()
	logrus.Info("Peripheral ID, name ", p.ID(), p.Name())
	p.Device().Connect(p)
}

func (c *Controller) onPeripheralConnected(p gatt.Peripheral, err error) {
	logrus.Printf("Connected to %s", p.Name())
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		logrus.Error(err, "Failed to set MTU")
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		logrus.Error(err, "Failed to discover services")
		return
	}

	for _, svc := range ss {

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, svc)
		if err != nil {
			logrus.Error(err, "Failed to discover characteristics")
			continue
		}

		for _, ch := range cs {
			property, found := findCharacteristic(c.spec, svc.UUID().String())
			if !found {
				continue
			}

			switch property.AccessMode {
			case v1alpha1.ReadOnly:
				{
					_, err := c.readCharacteristic(p, ch, property)
					if err != nil {
						logrus.Errorln(err)
						continue
					}
				}
			case v1alpha1.ReadWrite:
				{
					err := c.writeCharacteristic(p, ch, property)
					if err != nil {
						logrus.Errorln(err)
						return
					}
				}
			case v1alpha1.NotifyOnly:
				{
					err := c.getNotifyCharacteristic(p, ch, property)
					if err != nil {
						logrus.Errorln(err)
						return
					}
				}
			default:
				logrus.Errorf("AccessMode is not defined or either not a valid option, %s", property.AccessMode)
			}
		}
	}
	logrus.Info("Waiting for 5 seconds to get some notifications, if any.")
	time.Sleep(5 * time.Second)
}

func (c *Controller) onPeriphDisconnected(p gatt.Peripheral, err error) {
	logrus.Info("Device disconnected")
	if c.done != nil {
		close(c.done)
	}
}

func findCharacteristic(spec v1alpha1.BluetoothDeviceSpec, characteristicUUID string) (v1alpha1.DeviceProperty, bool) {
	deviceProperty := v1alpha1.DeviceProperty{}
	for _, p := range spec.Properties {
		if p.Visitor.CharacteristicUUID == characteristicUUID {
			return p, true
		}
	}
	return deviceProperty, false
}

func (c *Controller) readCharacteristic(p gatt.Peripheral, ch *gatt.Characteristic, property v1alpha1.DeviceProperty) (string, error) {
	b, err := p.ReadCharacteristic(ch)
	if err != nil {
		return "", err
	}
	logrus.Infof("ReadCharacteristic value %x | %q\n", b, b)

	convertedValue := fmt.Sprintf("%f", ConvertReadData(property, b))
	logrus.Infof("Converted read value to %s", convertedValue)

	c.updateDeviceStatus(property.Name, "", convertedValue)
	return convertedValue, nil
}

func (c *Controller) writeCharacteristic(p gatt.Peripheral, ch *gatt.Characteristic, property v1alpha1.DeviceProperty) error {
	if len(property.Visitor.DataWriteTo) == 0 {
		return fmt.Errorf("invalid length 0 of writeDataTo")
	}

	byteData, hasValue := findDataWriteToDeviceByDefaultValue(property.Visitor)
	if !hasValue {
		return fmt.Errorf("invalid length 0 of writeData")
	}

	err := p.WriteCharacteristic(ch, byteData, true)
	if err != nil {
		return fmt.Errorf("failed to write characteristic: %s with error: %s", ch.UUID(), err.Error())
	}

	value, err := c.readCharacteristic(p, ch, property)
	if err != nil {
		return fmt.Errorf("failed to read characteristic: %s with error: %s", ch.UUID(), err.Error())
	}
	c.updateDeviceStatus(property.Name, property.Visitor.DefaultValue, value)
	return nil
}

func (c *Controller) getNotifyCharacteristic(p gatt.Peripheral, ch *gatt.Characteristic, property v1alpha1.DeviceProperty) error {
	_, err := p.DiscoverDescriptors(nil, ch)
	if err != nil {
		return fmt.Errorf("failed to discover descriptors, %s", err.Error())
	}

	// Subscribe the characteristic, if possible.
	if (ch.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
		f := func(ch *gatt.Characteristic, b []byte, err error) {
			logrus.Infof("notified: % X | %q\n", b, b)
			value := fmt.Sprintf("%q", b)
			c.updateDeviceStatus(property.Name, "", value)
		}
		if err := p.SetNotifyValue(ch, f); err != nil {
			return fmt.Errorf("failed to subscribe characteristic, %s", err.Error())
		}
	}
	return nil
}

func findDataWriteToDeviceByDefaultValue(visitor v1alpha1.PropertyVisitor) ([]byte, bool) {
	for k, v := range visitor.DataWriteTo {
		if visitor.DefaultValue == k {
			return v, true
		}
	}
	return nil, false
}

func (c *Controller) updateDeviceStatus(name, desired, reported string) {
	sp := v1alpha1.StatusProperties{
		Name:      name,
		Desired:   desired,
		Reported:  reported,
		UpdatedAt: metav1.Time{Time: time.Now()},
	}
	found := false
	for i, property := range c.status.Properties {
		if property.Name == sp.Name {
			c.status.Properties[i] = sp
			found = true
			break
		}
	}
	if !found {
		c.status.Properties = append(c.status.Properties, sp)
	}
}
