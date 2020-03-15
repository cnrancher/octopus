package physical

import (
	"strconv"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/goburrow/modbus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

type Device interface {
	Configure(spec v1alpha1.ModbusDeviceSpec)
	On()
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler, modbusHandler modbus.ClientHandler, syncInterval time.Duration) Device {
	return &device{
		log:           log,
		name:          name,
		handler:       handler,
		modbusHandler: modbusHandler,
		syncInterval:  syncInterval,
	}
}

type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	status        v1alpha1.ModbusDeviceStatus
	modbusHandler modbus.ClientHandler
	syncInterval  time.Duration
}

func (d *device) Configure(spec v1alpha1.ModbusDeviceSpec) {
	properties := spec.Properties
	for _, property := range properties {
		if err := d.WriteProperty(property.DataType, property.Visitor, property.Value); err != nil {
			d.log.Error(err, "Error write property", "property", property.Name)
		}
	}
	d.UpdateStatus(properties)
}

func (d *device) On() {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	var ticker = time.NewTicker(d.syncInterval * time.Second)
	defer ticker.Stop()

	// periodically sync device status
	for {
		d.UpdateStatus(d.status.Properties)
		d.log.Info("Sync modbus device status", "device", d.name)
		select {
		case <-d.stop:
			return
		case <-ticker.C:
		}

	}
}

func (d *device) Shutdown() {
	if d.stop != nil {
		close(d.stop)
	}
	d.log.Info("Closed connection")
}

// write data of a property to coil register or holding register
func (d *device) WriteProperty(dataType v1alpha1.PropertyDataType, visitor v1alpha1.PropertyVisitor, value string) error {
	register := visitor.Register
	quantity := visitor.Quantity
	address := visitor.Offset

	client := modbus.NewClient(d.modbusHandler)
	switch register {
	case v1alpha1.ModbusRegisterTypeCoilRegister:
		// one bit per register
		l := quantity / 8
		if quantity%8 != 0 {
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
func (d *device) ReadProperty(dataType v1alpha1.PropertyDataType, visitor v1alpha1.PropertyVisitor) (string, error) {
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
	result, err = ByteArrayToString(data, dataType)
	if err != nil {
		d.log.Error(err, "Error converting to string", "datatype", dataType)
	}
	return result, nil
}

// update the properties from physical device to status
func (d *device) UpdateStatus(properties []v1alpha1.DeviceProperty) {
	d.Lock()
	defer d.Unlock()
	for i, property := range properties {
		value, err := d.ReadProperty(property.DataType, property.Visitor)
		if err != nil {
			d.log.Error(err, "Error sync device property", "property", property)
			continue
		}

		properties[i] = v1alpha1.DeviceProperty{Name: property.Name, DataType: property.DataType, Visitor: property.Visitor, Value: value}
	}
	d.status = v1alpha1.ModbusDeviceStatus{Properties: properties}
	d.handler(d.name, d.status)
}

func NewModbusHandler(config *v1alpha1.ModbusProtocolConfig, timeout time.Duration) (modbus.ClientHandler, error) {
	var TCPConfig = config.TCP
	var RTUConfig = config.RTU
	var handler modbus.ClientHandler

	if TCPConfig != nil {
		endpoint := TCPConfig.IP + ":" + strconv.Itoa(TCPConfig.Port)
		handlerTCP := modbus.NewTCPClientHandler(endpoint)
		handlerTCP.Timeout = timeout * time.Second
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
		handlerRTU.Timeout = timeout * time.Second
		if err := handlerRTU.Connect(); err != nil {
			return nil, err
		}
		defer handlerRTU.Close()
		handler = handlerRTU
	}
	return handler, nil
}
