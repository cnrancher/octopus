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

type modbusDevice struct {
	sync.Mutex

	log          logr.Logger
	instance     *v1alpha1.ModbusDevice
	toLimb       ModbusDeviceLimbSyncer
	stop         chan struct{}
	modbusClient *ModbusClient

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
	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) {
		if d.modbusClient != nil {
			if err := d.modbusClient.Close(); err != nil {
				if err != io.EOF {
					d.log.Error(err, "Error closing Modbus connection")
				}
			}
			d.modbusClient = nil
		}

		var modbusClient, err = NewModbusClient(newSpec.Protocol)
		if err != nil {
			return errors.Wrap(err, "failed to connect Modbus endpoint")
		}
		d.modbusClient = modbusClient

		// NB(thxCode) since the client has been changed,
		// we need to reset.
		d.instance.Spec = v1alpha1.ModbusDeviceSpec{}
	}

	return d.refresh(newSpec)
}

func (d *modbusDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.modbusClient != nil {
		if err := d.modbusClient.Close(); err != nil {
			if errors.Cause(err) != io.EOF {
				d.log.Error(err, "Error closing Modbus connection")
			}
		}
		d.modbusClient = nil
	}
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *modbusDevice) refresh(newSpec v1alpha1.ModbusDeviceSpec) error {
	var newStatus v1alpha1.ModbusDeviceStatus

	var staleStatus = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		var staleSpecPropsMap = mapSpecProperties(staleSpec.Properties)
		var staleStatusPropsMap = mapStatusProperties(staleStatus.Properties)

		// configures properties
		var statusProps = make([]v1alpha1.ModbusDeviceStatusProperty, 0, len(newSpec.Properties))
		for i := 0; i < len(newSpec.Properties); i++ {
			var specPropPtr = &newSpec.Properties[i]
			var statusProp v1alpha1.ModbusDeviceStatusProperty
			if staleStatusPropPtr, existed := staleStatusPropsMap[specPropPtr.Name]; existed {
				statusProp = *staleStatusPropPtr
			}

			for _, accessMode := range specPropPtr.MergeAccessModes() {
				switch accessMode {
				case v1alpha1.ModbusDevicePropertyAccessModeWriteOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				case v1alpha1.ModbusDevicePropertyAccessModeWriteMany:
					var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
					if err != nil {
						return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				case v1alpha1.ModbusDevicePropertyAccessModeReadOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.readProperty(specPropPtr)
						if err != nil {
							return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				default: // ModbusDevicePropertyAccessModeReadMany
					var statusPropPtr, err = d.readProperty(specPropPtr)
					if err != nil {
						return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				}
			}

			statusProps = append(statusProps, statusProp)
		}
		newStatus = v1alpha1.ModbusDeviceStatus{Properties: statusProps}
	} else {
		newStatus = staleStatus
	}

	// fetches in backend
	d.startFetch(newSpec.Protocol.GetSyncInterval())

	// records
	d.instance.Spec = newSpec
	d.instance.Status = newStatus
	return d.sync()
}

// writeProperty writes data of a property to CoilRegister or HoldingRegister.
func (d *modbusDevice) writeProperty(propPtr *v1alpha1.ModbusDeviceProperty, updatedAt *metav1.Time) (*v1alpha1.ModbusDeviceStatusProperty, error) {
	if propPtr.Value != "" {
		switch propPtr.Visitor.Register {
		case v1alpha1.ModbusDevicePropertyRegisterTypeCoilRegister:
			var err = write1BitRegister(propPtr, d.modbusClient.WriteCoils)
			if err != nil {
				return nil, err
			}
		case v1alpha1.ModbusDevicePropertyRegisterTypeHoldingRegister:
			var err = write16BitsRegister(propPtr, d.modbusClient.WriteHoldings)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("invalid writable register %s", propPtr.Visitor.Register)
		}
		d.log.V(4).Info("Write property", "property", propPtr.Name, "type", propPtr.Type, "value", propPtr.Value)
		updatedAt = now() // updates the timestamp
	}

	if updatedAt == nil {
		updatedAt = now() // records current timestamp
	}
	var statusPropPtr = &v1alpha1.ModbusDeviceStatusProperty{
		Name:        propPtr.Name,
		Type:        propPtr.Type,
		AccessModes: propPtr.AccessModes,
		UpdatedAt:   updatedAt,
	}
	return statusPropPtr, nil
}

// readProperty reads data of a property from its corresponding register.
// none arithmetic type is not supported to operate.
func (d *modbusDevice) readProperty(propPtr *v1alpha1.ModbusDeviceProperty) (*v1alpha1.ModbusDeviceStatusProperty, error) {
	var (
		value           string
		operationResult string
		err             error
	)

	switch propPtr.Visitor.Register {
	case v1alpha1.ModbusDevicePropertyRegisterTypeCoilRegister:
		value, operationResult, err = read1BitRegister(propPtr, d.modbusClient.ReadCoils)
		if err != nil {
			return nil, err
		}
	case v1alpha1.ModbusDevicePropertyRegisterTypeDiscreteInputRegister:
		value, operationResult, err = read1BitRegister(propPtr, d.modbusClient.ReadDiscreteInputs)
		if err != nil {
			return nil, err
		}
	case v1alpha1.ModbusDevicePropertyRegisterTypeHoldingRegister:
		value, operationResult, err = read16BitsRegister(propPtr, d.modbusClient.ReadHoldings)
		if err != nil {
			return nil, err
		}
	case v1alpha1.ModbusDevicePropertyRegisterTypeInputRegister:
		value, operationResult, err = read16BitsRegister(propPtr, d.modbusClient.ReadInputs)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("invalid readable register %s", propPtr.Visitor.Register)
	}
	d.log.V(4).Info("Read property", "property", propPtr.Name, "type", propPtr.Type, "value", value, "operationResult", operationResult)

	var statusPropPtr = &v1alpha1.ModbusDeviceStatusProperty{
		Name:            propPtr.Name,
		Type:            propPtr.Type,
		AccessModes:     propPtr.AccessModes,
		Value:           value,
		OperationResult: operationResult,
		UpdatedAt:       now(),
	}
	return statusPropPtr, nil
}

// fetch is blocked, it is used to sync the Modbus device status periodically,
// it's worth noting that it just reads or writes the "ReadMany/WriteMany" properties.
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

			// NB(thxCode) when the `spec.protocol` changes,
			// the `spec.properties` will be reset,
			// after obtaining the lock, this `fetch` goroutine should end.
			if len(d.instance.Status.Properties) != len(d.instance.Spec.Properties) {
				return
			}

			for i, statusProp := range d.instance.Status.Properties {
				var specPropPtr = &d.instance.Spec.Properties[i]

				for _, accessMode := range specPropPtr.MergeAccessModes() {
					switch accessMode {
					case v1alpha1.ModbusDevicePropertyAccessModeWriteMany:
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							// TODO give a way to feedback this to limb.
							d.log.Error(err, "Error for write property", "property", statusProp.Name)
							continue
						}
						statusProp = *statusPropPtr
					case v1alpha1.ModbusDevicePropertyAccessModeReadMany:
						var statusPropPtr, err = d.readProperty(specPropPtr)
						if err != nil {
							// TODO give a way to feedback this to limb.
							d.log.Error(err, "Error for read property", "property", statusProp.Name)
							continue
						}
						statusProp = *statusPropPtr
					default:
						continue
					}
				}

				d.instance.Status.Properties[i] = statusProp
			}
			if err := d.sync(); err != nil {
				d.log.Error(err, "Failed to sync")
			}
		}()

		select {
		case <-d.stop:
			return
		default:
		}
	}
}

// stopFetch stops the asynchronous fetch.
func (d *modbusDevice) stopFetch() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

// startFetch starts the asynchronous fetch.
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

type (
	// registerWriteFunc specifies the func to write single/multiple register.
	registerWriteFunc func(address, quantity uint16, value []byte) error

	// registerReadFunc specifies the func to read single/multiple register.
	registerReadFunc func(address, quantity uint16) ([]byte, error)
)

// write1BitRegister writes the given property's value to 1-bit register.
func write1BitRegister(prop *v1alpha1.ModbusDeviceProperty, write registerWriteFunc) error {
	var visitor = prop.Visitor
	var (
		data []byte
		err  error
	)

	/*
		single quantity
	*/

	if visitor.Quantity == 1 {
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString:
			// a 1-bit value can indicate by a size 2 hex string.
			if len(prop.Value) != 2 {
				return errors.Errorf("the length of single 1-bit quantity %s's hex value is invalid, expected is 2", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeBinaryString:
			// a 1-bit value can indicate by a size 8 binary string.
			if len(prop.Value) != 8 {
				return errors.Errorf("the length of single 1-bit quantity %s's binary value is invalid, expected is 8", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeBase64String:
			// padding base64 string
			if len(prop.Value) != 4 {
				return errors.Errorf("the length of single 1-bit quantity %s's base64 value is invalid, expected is 4", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeBoolean:
			// pass
		default:
			return errors.Errorf("single 1-bit quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		data, err = convertValueToBytes(prop)
		if err != nil {
			return errors.Errorf("failed to convert the single 1-bit quantity %s's value to %s type", visitor.Register, prop.Type)
		}

		err = write(visitor.Offset, visitor.Quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s type %s value to single 1-bit quantity %s", prop.Type, prop.Value, visitor.Register)
		}
		return nil
	}

	/*
		multiple quantities
	*/

	// NB(thxCode) multiple quantities CoilRegister can only be considered as hexString/binaryString type,
	// for example, if we write 20# - 29# CoilRegisters at the same time,
	// which means we can input "CD01" to indicate the following multiple CoilRegister values:
	// coil num      : 27 26 25 24 23 22 21 20  -  -  -  -  -  - 29 28
	// binary value  :  1  1  0  0  1  1  0  1  0  0  0  0  0  0  0  1|  <-- input/output
	// hex value     :  *  *  *  C| *  *  *  D| *  *  *  0| *  *  *  1|  <-- input/output
	// byte value    :  *  *  *  *  *  *  *205| *  *  *  *  *  *  *  1|  <-- convert
	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeHexString:
		// NB(thxCode) 8 quantity == 1 byte = 2 hex chars
		if s := ((visitor.Quantity + 7) >> 3) << 1; s != uint16(len(prop.Value)) {
			return errors.Errorf("the length of multiple 1-bit quantities %s's hex value is invalid, expected is %d", visitor.Register, s)
		}
	case v1alpha1.ModbusDevicePropertyTypeBinaryString:
		// NB(thxCode) 8 quantity == 1 byte = 8 binary chars
		if s := ((visitor.Quantity + 7) >> 3) << 3; s != uint16(len(prop.Value)) {
			return errors.Errorf("the length of multiple 1-bit quantities %s's binary value is invalid, expected is %d", visitor.Register, s)
		}
	case v1alpha1.ModbusDevicePropertyTypeBase64String:
		// NB(thxCode) 8 quantity == 1 byte, 3 bytes = 4 base64 chars
		if s := (((visitor.Quantity + 7) >> 3) * 3) >> 2; s != uint16(len(prop.Value)) {
			return errors.Errorf("the length of multiple 1-bit quantities %s's base64 value is invalid, expected is %d", visitor.Register, s)
		}
		// pass
	default:
		return errors.Errorf("multiple 1-bit quantities %s can only set as hexString or binaryString type", visitor.Register)
	}

	data, err = convertValueToBytes(prop)
	if err != nil {
		return errors.Errorf("failed to convert the multiple 1-bit quantities %s's value to %s type", visitor.Register, prop.Type)
	}

	err = write(visitor.Offset, visitor.Quantity, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s type %s value to multiple 1-bit quantities %s", prop.Type, prop.Value, visitor.Register)
	}
	return nil
}

// write16BitsRegister writes the given property's value to 16-bits register.
func write16BitsRegister(prop *v1alpha1.ModbusDeviceProperty, write registerWriteFunc) error {
	var visitor = prop.Visitor
	var (
		data []byte
		err  error
	)

	/*
		single quantity
	*/

	if visitor.Quantity == 1 {
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString:
			// NB(thxCode) 1 quantity = 2 bytes = 4 hex chars
			// a 16-bits value can indicate by a size 4 hex string.
			if len(prop.Value) != 4 {
				return errors.Errorf("the length of single 16-bits quantity %s's hex value is invalid, expected is 4", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeBinaryString:
			// NB(thxCode) 1 quantity = 2 bytes = 16 binary chars
			// a 16-bits value can indicate by a size 16 binary string.
			if len(prop.Value) != 16 {
				return errors.Errorf("the length of single 16-bits quantity %s's binary value is invalid, expected is 16", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeBase64String:
			// padding base64 string
			if len(prop.Value) != 4 {
				return errors.Errorf("the length of single 16-bits quantity %s's base64 value is invalid, expected is 4", visitor.Register)
			}
		case v1alpha1.ModbusDevicePropertyTypeInt8,
			v1alpha1.ModbusDevicePropertyTypeUint8:
			// pass
		case v1alpha1.ModbusDevicePropertyTypeInt16,
			v1alpha1.ModbusDevicePropertyTypeUint16:
			// pass
		default:
			return errors.Errorf("single 16-bits quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		data, err = convertValueToBytes(prop)
		if err != nil {
			return errors.Errorf("failed to convert the single 16-bits quantity %s's value to %s type", visitor.Register, prop.Type)
		}

		err = write(visitor.Offset, visitor.Quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s type %s value to single 16-bits quantity %s", prop.Type, prop.Value, visitor.Register)
		}
		return nil
	}

	/*
		multiple quantities
	*/

	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt32,
		v1alpha1.ModbusDevicePropertyTypeInt,
		v1alpha1.ModbusDevicePropertyTypeUint32,
		v1alpha1.ModbusDevicePropertyTypeUint:
		if visitor.Quantity != 2 {
			return errors.Errorf("multiple 16-bits quantities %s cannot set as int/int32/uint/uint32 type, the amount of quantity is invalid, it must be 2", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeInt64,
		v1alpha1.ModbusDevicePropertyTypeUint64:
		if visitor.Quantity != 4 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as int64/uint64 type, the amount of quantity is invalid, it must be 4", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeFloat32,
		v1alpha1.ModbusDevicePropertyTypeFloat:
		if visitor.Quantity != 2 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as float/float32 type, the amount of quantity is invalid, it must be 2", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeFloat64,
		v1alpha1.ModbusDevicePropertyTypeDouble:
		if visitor.Quantity != 4 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as double/float64 type, the amount of quantity is invalid, it must be 4", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeHexString:
		// NB(thxCode) 1 quantity = 2 bytes = 4 hex chars
		// multiple 16-bits value can indicate by a size 4N hex string.
		if visitor.Quantity != uint16(len(prop.Value)>>2) {
			return errors.Errorf("the length of multiple 16-bits quantities %s's hex value is invalid, expected is 4 * quantity", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeBinaryString:
		// NB(thxCode) 1 quantity = 2 bytes = 16 binary chars
		// multiple 16-bits value can indicate by a size 16N binary string.
		if visitor.Quantity != uint16(len(prop.Value)>>4) {
			return errors.Errorf("the length of multiple 16-bits quantities %s's binary value is invalid, expected is 16 * quantity", visitor.Register)
		}
	case v1alpha1.ModbusDevicePropertyTypeBase64String:
		// NB(thxCode) 1 quantity = 2 bytes, 3 bytes = 4 base64 chars
		// multiple 16-bits value can indicate by a size 3/8N base64 string.
		if (visitor.Quantity+2)/3 != uint16(len(prop.Value)>>3) {
			return errors.Errorf("the length of multiple 16-bits quantities %s's base64 value is invalid, expected is 8*⌈quantity/3⌉", visitor.Register)
		}
		// pass
	case v1alpha1.ModbusDevicePropertyTypeString:
		// NB(thxCode) 1 quantity = 2 bytes, 4 bytes = 1 rune
		// we don't need to verify prop.Value's length in here,
		// we can truncate the bytes below the visitor.Quantity * 2.
		if visitor.Quantity%2 != 0 {
			return errors.Errorf("the amount of multiple 16-bits quantities %s's quantity is invalid, expected is 2N", visitor.Register)
		}
	default:
		return errors.Errorf("multiple 16-bits quantities %s cannot set as %s type", visitor.Register, prop.Type)
	}

	data, err = convertValueToBytes(prop)
	if err != nil {
		return errors.Errorf("failed to convert the multiple 16-bits quantities %s's value to %s type", visitor.Register, prop.Type)
	}

	err = write(visitor.Offset, visitor.Quantity, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s type %s value to multiple 16-bits quantities %s", prop.Type, prop.Value, visitor.Register)
	}
	return nil
}

// read1BitRegister reads the value of given property from 1-bit register.
func read1BitRegister(prop *v1alpha1.ModbusDeviceProperty, read registerReadFunc) (string, string, error) {
	var visitor = prop.Visitor

	/*
		single quantity
	*/

	if visitor.Quantity == 1 {
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeBase64String,
			v1alpha1.ModbusDevicePropertyTypeHexString,
			v1alpha1.ModbusDevicePropertyTypeBinaryString:
			// pass
		case v1alpha1.ModbusDevicePropertyTypeBoolean:
			// pass
		default:
			return "", "", errors.Errorf("single 1-bit quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		var val, err = read(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to read type %s bytes from single 1-bit quantity %s", prop.Type, visitor.Register)
		}
		if len(val) != 1 {
			return "", "", errors.Errorf("read type %s bytes from single 1-bit quantity %s, but the size of bytes is invalid", prop.Type, visitor.Register)
		}

		return parseValueFromBytes(val, prop)
	}

	/*
		multiple quantities
	*/

	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeBase64String,
		v1alpha1.ModbusDevicePropertyTypeHexString,
		v1alpha1.ModbusDevicePropertyTypeBinaryString:
		// pass
	default:
		return "", "", errors.Errorf("multiple 1-bit quantities %s can only set as hexString or binaryString type", visitor.Register)
	}

	var val, err = read(visitor.Offset, visitor.Quantity)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to read type %s bytes from multiple 1-bit quantities %s", prop.Type, visitor.Register)
	}
	if len(val) <= 1 {
		return "", "", errors.Errorf("failed to read type %s bytes from multiple 1-bits quantities %s, but the size of bytes is invalid", prop.Type, visitor.Register)
	}

	return parseValueFromBytes(val, prop)
}

// read16BitsRegister reads the value of given property from 16-bits register.
func read16BitsRegister(prop *v1alpha1.ModbusDeviceProperty, read registerReadFunc) (string, string, error) {
	var visitor = prop.Visitor

	/*
		single quantity
	*/

	if visitor.Quantity == 1 {
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString,
			v1alpha1.ModbusDevicePropertyTypeBinaryString,
			v1alpha1.ModbusDevicePropertyTypeBase64String:
			// pass
		case v1alpha1.ModbusDevicePropertyTypeInt8, v1alpha1.ModbusDevicePropertyTypeUint8:
			// pass
		case v1alpha1.ModbusDevicePropertyTypeInt16, v1alpha1.ModbusDevicePropertyTypeUint16:
			// pass
		default:
			return "", "", errors.Errorf("single 16-bits quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		var val, err = read(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to read type %s bytes from single 16-bits quantity %s", prop.Type, visitor.Register)
		}
		if len(val) != 2 {
			return "", "", errors.Errorf("failed to read type %s bytes from single 16-bits quantity %s, but the size is invalid", prop.Type, visitor.Register)
		}

		return parseValueFromBytes(val, prop)
	}

	/*
		multiple quantities
	*/

	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeHexString, v1alpha1.ModbusDevicePropertyTypeBinaryString,
		v1alpha1.ModbusDevicePropertyTypeBase64String, v1alpha1.ModbusDevicePropertyTypeString:
		// pass
	case v1alpha1.ModbusDevicePropertyTypeInt, v1alpha1.ModbusDevicePropertyTypeUint,
		v1alpha1.ModbusDevicePropertyTypeInt32, v1alpha1.ModbusDevicePropertyTypeUint32,
		v1alpha1.ModbusDevicePropertyTypeInt64, v1alpha1.ModbusDevicePropertyTypeUint64:
		// pass
	case v1alpha1.ModbusDevicePropertyTypeFloat, v1alpha1.ModbusDevicePropertyTypeFloat32,
		v1alpha1.ModbusDevicePropertyTypeDouble, v1alpha1.ModbusDevicePropertyTypeFloat64:
		// pass
	default:
		return "", "", errors.Errorf("multiple 16-bits quantities %s cannot set as %s type", visitor.Register, prop.Type)
	}

	var val, err = read(visitor.Offset, visitor.Quantity)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to read type %s bytes from multiple quantities %s", prop.Type, visitor.Register)
	}
	if len(val) <= 2 {
		return "", "", errors.Errorf("failed to read type %s bytes from multiple 16-bits quantities %s, but the size is invalid", prop.Type, visitor.Register)
	}

	return parseValueFromBytes(val, prop)
}

func mapSpecProperties(specProps []v1alpha1.ModbusDeviceProperty) map[string]*v1alpha1.ModbusDeviceProperty {
	var ret = make(map[string]*v1alpha1.ModbusDeviceProperty, len(specProps))
	for i := 0; i < len(specProps); i++ {
		var prop = specProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func mapStatusProperties(statusProps []v1alpha1.ModbusDeviceStatusProperty) map[string]*v1alpha1.ModbusDeviceStatusProperty {
	var ret = make(map[string]*v1alpha1.ModbusDeviceStatusProperty, len(statusProps))
	for i := 0; i < len(statusProps); i++ {
		var prop = statusProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}
