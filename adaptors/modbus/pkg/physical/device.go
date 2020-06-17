package physical

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Device interface {
	Configure(references api.ReferencesHandler, obj v1alpha1.ModbusDevice) error
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler) Device {
	return &device{
		log:     log,
		name:    name,
		handler: handler,
	}
}

const (
	mqttTimeout = 5 * time.Second
	bits        = 8
)

type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	spec   v1alpha1.ModbusDeviceSpec
	status v1alpha1.ModbusDeviceStatus

	modbusHandler modbus.ClientHandler
	mqttClient    mqtt.Client
}

func (d *device) Configure(references api.ReferencesHandler, obj v1alpha1.ModbusDevice) error {
	spec := d.spec
	d.spec = obj.Spec

	// configure protocol config and parameters
	if !reflect.DeepEqual(d.spec.ProtocolConfig, spec.ProtocolConfig) || !reflect.DeepEqual(d.spec.Parameters, spec.Parameters) {
		var modbusHandler, err = newModbusHandler(d.spec.ProtocolConfig, d.spec.Parameters.Timeout.Duration)
		d.modbusHandler = modbusHandler
		if err != nil {
			d.log.Error(err, "Failed to connect to modbus device endpoint")
			return err
		}
		// if connected and sync interval changed, reconfigure sync interval
		d.on()
	}

	if !reflect.DeepEqual(d.spec.Extension, spec.Extension) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect(mqttTimeout)
			d.mqttClient = nil

			// since there is only a MQTT inside extension field, here can set to nil directly.
			d.status.Extension = nil
		}

		if d.spec.Extension.MQTT != nil {
			var cli, outline, err = mqtt.NewClient(&obj, *d.spec.Extension.MQTT, references.ToDataMap())
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT client")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli

			if d.status.Extension == nil {
				d.status.Extension = &v1alpha1.DeviceExtensionStatus{}
			}
			d.status.Extension.MQTT = outline
		}
	}

	// configure properties
	for _, property := range d.spec.Properties {
		if property.ReadOnly {
			continue
		}
		if err := d.writeProperty(property.DataType, property.Visitor, property.Value); err != nil {
			d.log.Error(err, "Error write property", "property", property)
			continue
		}
		d.log.Info("Write property", "property", property)
	}
	d.updateStatus(d.spec.Properties)
	return nil
}

func (d *device) on() {
	// close connection to old device
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	d.log.Info("Connect to device", "device", d.name)

	// periodically sync device status
	go func() {
		var ticker = time.NewTicker(d.spec.Parameters.SyncInterval.Duration)
		defer ticker.Stop()

		for {
			d.updateStatus(d.spec.Properties)
			d.log.Info("Sync modbus device status", "properties", d.status.Properties)
			select {
			case <-d.stop:
				return
			case <-ticker.C:
			}
		}
	}()
}

func (d *device) Shutdown() {
	if d.stop != nil {
		close(d.stop)
	}

	if d.mqttClient != nil {
		d.mqttClient.Disconnect(mqttTimeout)
		d.mqttClient = nil
	}

	d.log.Info("Closed connection")
}

// write data of a property to coil register or holding register
func (d *device) writeProperty(dataType v1alpha1.PropertyDataType, visitor v1alpha1.PropertyVisitor, value string) error {
	register := visitor.Register
	quantity := visitor.Quantity
	address := visitor.Offset

	client := modbus.NewClient(d.modbusHandler)
	switch register {
	case v1alpha1.ModbusRegisterTypeCoilRegister:
		// one bit per register
		l := quantity / bits
		if quantity%bits != 0 {
			l++
		}
		data, err := StringToByteArray(value, dataType, int(l))
		if err != nil {
			d.log.Error(err, "Error converting data to byte array", "value", value)
			return err
		}
		_, err = client.WriteMultipleCoils(address, quantity, data)
		if err != nil {
			d.log.Error(err, "Error writing property to register", "register", register, "data", data)
			return err
		}
	case v1alpha1.ModbusRegisterTypeHoldingRegister:
		// two bytes per register
		data, err := StringToByteArray(value, dataType, int(quantity*2))
		if err != nil {
			d.log.Error(err, "Error converting data to byte array", "value", value)
			return err
		}
		_, err = client.WriteMultipleRegisters(address, quantity, data)
		if err != nil {
			d.log.Error(err, "Error writing property to register", "register", register, "data", data)
			return err
		}
	}
	return nil
}

// read data of a property from its corresponding register
func (d *device) readProperty(dataType v1alpha1.PropertyDataType, visitor v1alpha1.PropertyVisitor) (string, error) {
	register := visitor.Register
	quantity := visitor.Quantity
	address := visitor.Offset

	var result string
	var data []byte
	var err error
	client := modbus.NewClient(d.modbusHandler)
	switch register {
	case v1alpha1.ModbusRegisterTypeCoilRegister:
		data, err = client.ReadCoils(address, quantity)
		if err != nil {
			d.log.Error(err, "Error reading property from register", "register", register)
			return "", err
		}
	case v1alpha1.ModbusRegisterTypeDiscreteInputRegister:
		data, err = client.ReadDiscreteInputs(address, quantity)
		if err != nil {
			d.log.Error(err, "Error reading property from register", "register", register)
			return "", err
		}

	case v1alpha1.ModbusRegisterTypeHoldingRegister:
		data, err = client.ReadHoldingRegisters(address, quantity)
		if err != nil {
			d.log.Error(err, "Error reading property from register", "register", register)
			return "", err
		}

	case v1alpha1.ModbusRegisterTypeInputRegister:
		data, err = client.ReadInputRegisters(address, quantity)
		if err != nil {
			d.log.Error(err, "Error reading property from register", "register", register)
			return "", err
		}

	}
	result, err = ByteArrayToString(data, dataType, visitor.OrderOfOperations)
	if err != nil {
		d.log.Error(err, "Error converting to string", "datatype", dataType)
	}
	return result, nil
}

// update the properties from physical device to status
func (d *device) updateStatus(properties []v1alpha1.DeviceProperty) {
	d.Lock()
	defer d.Unlock()
	for _, property := range properties {
		value, err := d.readProperty(property.DataType, property.Visitor)
		if err != nil {
			d.log.Error(err, "Error sync device property", "property", property)
			continue
		}
		d.updateStatusProperty(property.Name, value, property.DataType)
	}
	d.handler(d.name, d.status)

	if d.mqttClient != nil {
		var status = d.status.DeepCopy()
		status.Extension = nil
		if err := d.mqttClient.Publish(status); err != nil {
			d.log.Error(err, "Failed to publish MQTT message")
		}
	}
}

func (d *device) updateStatusProperty(name, value string, dataType v1alpha1.PropertyDataType) {
	sp := v1alpha1.StatusProperties{
		Name:      name,
		Value:     value,
		DataType:  dataType,
		UpdatedAt: metav1.Time{Time: time.Now()},
	}
	found := false
	for i, property := range d.status.Properties {
		if property.Name == sp.Name {
			d.status.Properties[i] = sp
			found = true
			break
		}
	}
	if !found {
		d.status.Properties = append(d.status.Properties, sp)
	}
}
func newModbusHandler(config *v1alpha1.ModbusProtocolConfig, timeout time.Duration) (modbus.ClientHandler, error) {
	var TCPConfig = config.TCP
	var RTUConfig = config.RTU
	var handler modbus.ClientHandler

	if TCPConfig != nil {
		endpoint := TCPConfig.IP + ":" + strconv.Itoa(TCPConfig.Port)
		handlerTCP := modbus.NewTCPClientHandler(endpoint)
		handlerTCP.Timeout = timeout
		handlerTCP.SlaveId = byte(TCPConfig.SlaveID)
		if err := handlerTCP.Connect(); err != nil {
			return nil, err
		}
		defer handlerTCP.Close()
		handler = handlerTCP
	} else if RTUConfig != nil {
		serialPort := RTUConfig.SerialPort
		handlerRTU := modbus.NewRTUClientHandler(serialPort)
		handlerRTU.BaudRate = RTUConfig.BaudRate
		handlerRTU.DataBits = RTUConfig.DataBits
		handlerRTU.Parity = RTUConfig.Parity
		handlerRTU.StopBits = RTUConfig.StopBits
		handlerRTU.SlaveId = byte(RTUConfig.SlaveID)
		handlerRTU.Timeout = timeout
		if err := handlerRTU.Connect(); err != nil {
			return nil, err
		}
		defer handlerRTU.Close()
		handler = handlerRTU
	}
	return handler, nil
}
