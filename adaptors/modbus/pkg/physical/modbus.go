package physical

import (
	"io"
	stdlog "log"
	"os"

	"github.com/go-logr/logr"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/util/log/logflag"
)

// ModbusClosableClientHandler combines modbus.ClientHandler and io.Closer interfaces.
type ModbusClosableClientHandler interface {
	modbus.ClientHandler
	io.Closer
}

// ModbusClient is a wrapper to handle the real modbus.Client.
type ModbusClient struct {
	log     logr.Logger
	handler ModbusClosableClientHandler
}

// WriteCoils writes value to CoilRegisters.
func (c *ModbusClient) WriteCoils(address, quantity uint16, value []byte) error {
	var _, err = c.rawClient().WriteMultipleCoils(address, quantity, value)
	c.log.V(4).Info("Write coil registers")
	return err
}

// WriteHoldings writes value to HoldingRegisters.
func (c *ModbusClient) WriteHoldings(address, quantity uint16, value []byte) error {
	var _, err = c.rawClient().WriteMultipleRegisters(address, quantity, value)
	c.log.V(4).Info("Write holding registers")
	return err
}

// ReadCoils reads value from CoilRegisters.
func (c *ModbusClient) ReadCoils(address, quantity uint16) ([]byte, error) {
	var data, err = c.rawClient().ReadCoils(address, quantity)
	c.log.V(4).Info("Read coli registers")
	return data, err
}

// ReadDiscreteInputs reads value from DiscreteInputRegisters.
func (c *ModbusClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
	var data, err = c.rawClient().ReadDiscreteInputs(address, quantity)
	c.log.V(4).Info("Read discrete input registers")
	return data, err
}

// ReadHoldings reads value from HoldingRegisters.
func (c *ModbusClient) ReadHoldings(address, quantity uint16) ([]byte, error) {
	var data, err = c.rawClient().ReadHoldingRegisters(address, quantity)
	c.log.V(4).Info("Read holding registers")
	return data, err
}

// ReadInputs reads value from InputRegisters.
func (c *ModbusClient) ReadInputs(address, quantity uint16) ([]byte, error) {
	var data, err = c.rawClient().ReadInputRegisters(address, quantity)
	c.log.V(4).Info("Read input registers")
	return data, err
}

func (c *ModbusClient) Close() error {
	var err = c.handler.Close()
	c.log.V(4).Info("Closed")
	return err
}

// rawClient returns the raw modbus.Client.
func (c *ModbusClient) rawClient() modbus.Client {
	return modbus.NewClient(c.handler)
}

// NewModbusClient creates Modbus client.
func NewModbusClient(protocol v1alpha1.ModbusDeviceProtocol) (*ModbusClient, error) {
	var stdLogger *stdlog.Logger
	if logflag.GetLogVerbosity() > 3 {
		stdLogger = stdlog.New(os.Stdout, "modbus.underlay.client", stdlog.LstdFlags)
	}

	if protocol.TCP != nil {
		var tcpConfig = protocol.TCP
		var tcpClientHandler = modbus.NewTCPClientHandler(tcpConfig.Endpoint)
		tcpClientHandler.Timeout = tcpConfig.GetConnectTimeout()
		tcpClientHandler.IdleTimeout = tcpConfig.GetSyncInterval() * 3
		tcpClientHandler.SlaveId = byte(tcpConfig.WorkerID)
		tcpClientHandler.Logger = stdLogger
		if err := tcpClientHandler.Connect(); err != nil {
			return nil, errors.Wrap(err, "failed to connect via TCP")
		}

		var logger = log.WithName("modbus.client").WithValues("protocol", "tcp", "endpoint", tcpConfig.Endpoint)
		return &ModbusClient{log: logger, handler: tcpClientHandler}, nil
	}

	if protocol.RTU != nil {
		var rtuConfig = protocol.RTU
		rtuClientHandler := modbus.NewRTUClientHandler(rtuConfig.Endpoint)
		rtuClientHandler.BaudRate = rtuConfig.BaudRate
		rtuClientHandler.DataBits = rtuConfig.DataBits
		rtuClientHandler.Parity = rtuConfig.Parity
		rtuClientHandler.StopBits = rtuConfig.StopBits
		rtuClientHandler.SlaveId = byte(rtuConfig.WorkerID)
		rtuClientHandler.Timeout = rtuConfig.GetConnectTimeout()
		rtuClientHandler.IdleTimeout = rtuConfig.GetSyncInterval() * 3
		rtuClientHandler.Logger = stdLogger
		if err := rtuClientHandler.Connect(); err != nil {
			return nil, errors.Wrap(err, "failed to connect via RTU")
		}

		var logger = log.WithName("modbus.client").WithValues("protocol", "rtu", "endpoint", rtuConfig.Endpoint)
		return &ModbusClient{log: logger, handler: rtuClientHandler}, nil
	}

	return nil, errors.New("failed to create Modbus handler with empty protocol")
}
