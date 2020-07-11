package physical

import (
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/modbus/pkg/metadata"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/object"
)

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, device *v1alpha1.ModbusDevice) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb ModbusDeviceLimbSyncer) Device {
	log.Info("Created ")
	return &modbusDevice{
		log: log,
		instance: &v1alpha1.ModbusDevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

const (
	bits = 8
)

type modbusDevice struct {
	sync.Mutex

	log           logr.Logger
	instance      *v1alpha1.ModbusDevice
	toLimb        ModbusDeviceLimbSyncer
	stop          chan struct{}
	modbusHandler ModbusClientHandler

	mqttClient mqtt.Client
}

func (d *modbusDevice) Configure(references api.ReferencesHandler, device *v1alpha1.ModbusDevice) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec
	var staleSpec = d.instance.Spec

	// configures MQTT client if needed
	var staleExtension, newExtension v1alpha1.ModbusDeviceExtension
	if staleSpec.Extension != nil {
		staleExtension = *staleSpec.Extension
	}
	if newSpec.Extension != nil {
		newExtension = *newSpec.Extension
	}
	if !reflect.DeepEqual(staleExtension.MQTT, newExtension.MQTT) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
		}

		if newExtension.MQTT != nil {
			var cli, err = mqtt.NewClient(*newExtension.MQTT, object.GetControlledOwnerObjectReference(device), references)
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT client")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli
		}
	}

	// configures Modbus client
	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) || !reflect.DeepEqual(staleSpec.Parameters, newSpec.Parameters) {
		if d.modbusHandler != nil {
			if err := d.modbusHandler.Close(); err != nil {
				if err != io.EOF {
					d.log.Error(err, "Error closing Modbus connection")
				}
			}
			d.modbusHandler = nil
		}

		var handler, err = NewModbusClientHandler(newSpec.Protocol, newSpec.Parameters.GetTimeout())
		if err != nil {
			return errors.Wrap(err, "failed to connect Modbus endpoint")
		}
		d.modbusHandler = handler
	}

	return d.refresh(newSpec)
}

func (d *modbusDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.modbusHandler != nil {
		if err := d.modbusHandler.Close(); err != nil {
			if err != io.EOF {
				d.log.Error(err, "Error closing Modbus connection")
			}
		}
		d.modbusHandler = nil
	}
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *modbusDevice) refresh(newSpec v1alpha1.ModbusDeviceSpec) error {
	var status = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		// configures properties
		var specProps = newSpec.Properties
		var statusProps = make([]v1alpha1.ModbusDeviceStatusProperty, 0, len(specProps))
		for _, prop := range specProps {
			if !prop.ReadOnly {
				if err := d.writeProperty(prop.Type, prop.Visitor, prop.Value); err != nil {
					return errors.Wrapf(err, "failed to write property %s", prop.Name)
				}
				d.log.V(4).Info("Write property", "property", prop.Name, "type", prop.Type)
			}
			value, err := d.readProperty(prop.Type, prop.Visitor)
			if err != nil {
				return errors.Wrapf(err, "failed to read property %s", prop.Name)
			}
			d.log.V(4).Info("Read property", "property", prop.Name, "type", prop.Type)
			statusProps = append(statusProps, v1alpha1.ModbusDeviceStatusProperty{
				Name:      prop.Name,
				Value:     value,
				Type:      prop.Type,
				UpdatedAt: now(),
			})
		}
		status = v1alpha1.ModbusDeviceStatus{Properties: statusProps}
	}

	// fetches in backend
	d.startFetch(newSpec.Parameters.GetSyncInterval())

	// records
	d.instance.Spec = newSpec
	d.instance.Status = status
	return d.sync()
}

// fetch is blocked, it is used to sync the modbus device status periodically,
// it's worth noting that it just reads the properties from modbus device.
func (d *modbusDevice) fetch(interval time.Duration, stop <-chan struct{}) {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.log.Info("Fetching")
	defer func() {
		d.log.Info("Finished fetching")
	}()

	var ticker = time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		d.Lock()
		func() {
			defer d.Unlock()

			// read according to the properties defined by the spec,
			// and finally fill it back to status.
			var specProps = d.instance.Spec.Properties
			var statusProps = make([]v1alpha1.ModbusDeviceStatusProperty, 0, len(specProps))
			for _, prop := range specProps {
				var value, err = d.readProperty(prop.Type, prop.Visitor)
				if err != nil {
					// TODO give a way to feedback this to limb.
					d.log.Error(err, "Error fetching device property", "property", prop.Name)
				}
				d.log.V(4).Info("Read property", "property", prop.Name, "type", prop.Type)
				statusProps = append(statusProps, v1alpha1.ModbusDeviceStatusProperty{
					Name:      prop.Name,
					Value:     value,
					Type:      prop.Type,
					UpdatedAt: now(),
				})
			}
			d.instance.Status.Properties = statusProps
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}()

		select {
		case <-d.stop:
			return
		default:
		}
	}
}

// writeProperty writes data of a property to CoilRegister or HoldingRegister.
func (d *modbusDevice) writeProperty(dataType v1alpha1.ModbusDevicePropertyType, visitor v1alpha1.ModbusDevicePropertyVisitor, value string) error {
	var client = d.modbusHandler.Connect()

	// NB(thxCode) don't write the property if the value is blank.
	if value == "" {
		return nil
	}

	switch visitor.Register {
	case v1alpha1.ModbusDeviceCoilRegister:
		// one bit per register
		var quantity = visitor.Quantity
		var length = quantity / bits
		if quantity%bits != 0 {
			length++
		}

		var data, err = StringToByteArray(value, dataType, int(length))
		if err != nil {
			return errors.Wrapf(err, "failed to convert %s string to %s byte array", value, dataType)
		}
		_, err = client.WriteMultipleCoils(visitor.Offset, quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s to %s register", data, visitor.Register)
		}
	case v1alpha1.ModbusDeviceHoldingRegister:
		// two bytes per register
		var quantity = visitor.Quantity
		var length = quantity * 2

		var data, err = StringToByteArray(value, dataType, int(length))
		if err != nil {
			return errors.Wrapf(err, "failed to convert %s string to %s byte array", value, dataType)
		}
		_, err = client.WriteMultipleRegisters(visitor.Offset, quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s to %s register", data, visitor.Register)
		}
	}
	return nil
}

// readProperty reads data of a property from its corresponding register.
func (d *modbusDevice) readProperty(dataType v1alpha1.ModbusDevicePropertyType, visitor v1alpha1.ModbusDevicePropertyVisitor) (result string, err error) {
	var client = d.modbusHandler.Connect()

	var data []byte
	switch visitor.Register {
	case v1alpha1.ModbusDeviceCoilRegister:
		data, err = client.ReadCoils(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read property from %s register", visitor.Register)
		}
	case v1alpha1.ModbusDeviceDiscreteInputRegister:
		data, err = client.ReadDiscreteInputs(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read property from %s register", visitor.Register)
		}
	case v1alpha1.ModbusDeviceHoldingRegister:
		data, err = client.ReadHoldingRegisters(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read property from %s register", visitor.Register)
		}
	case v1alpha1.ModbusDeviceInputRegister:
		data, err = client.ReadInputRegisters(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read property from %s register", visitor.Register)
		}
	default:
		return "", errors.Errorf("invalid readable register %s", visitor.Register)
	}

	result, err = ByteArrayToString(data, dataType, visitor.OrderOfOperations)
	if err != nil {
		return "", errors.Wrapf(err, "failed to convert %s byte array to string", dataType)
	}
	return result, nil
}

func (d *modbusDevice) stopFetch() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

func (d *modbusDevice) startFetch(fetchInterval time.Duration) {
	if d.stop == nil {
		d.stop = make(chan struct{})
		go d.fetch(fetchInterval, d.stop)
	}
}

// sync combines all synchronization operations.
func (d *modbusDevice) sync() error {
	if d.toLimb != nil {
		if err := d.toLimb(d.instance); err != nil {
			return err
		}
	}
	if d.mqttClient != nil {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: d.instance.Status}); err != nil {
			return err
		}
	}
	d.log.V(1).Info("Synced")
	return nil
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}
