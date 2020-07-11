package physical

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/goburrow/modbus"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/log/logflag"
)

// ModbusDeviceLimbSyncer is used to sync modebus device to limb.
type ModbusDeviceLimbSyncer func(in *v1alpha1.ModbusDevice) error

// ModbusClientHandler is a wrapper for handling modbus.ClientHandler and modbus.Client.
type ModbusClientHandler interface {
	io.Closer

	// Connect returns a modbus.Client
	Connect() modbus.Client
}

type modbusClientHandler struct {
	handler modbus.ClientHandler
	closer  io.Closer
}

func (h *modbusClientHandler) Connect() modbus.Client {
	return modbus.NewClient(h.handler)
}

func (h *modbusClientHandler) Close() error {
	return h.closer.Close()
}

// NewModbusClientHandler creates a ModbusClientHandler.
func NewModbusClientHandler(protocol v1alpha1.ModbusDeviceProtocol, timeout time.Duration) (ModbusClientHandler, error) {
	var logger *log.Logger
	if logflag.GetLogVerbosity() > 4 {
		logger = log.New(os.Stdout, "modbus.client", log.LstdFlags)
	}

	if protocol.TCP != nil {
		var tcpConfig = protocol.TCP
		var tcpClientHandler = modbus.NewTCPClientHandler(tcpConfig.Endpoint)
		tcpClientHandler.Timeout = timeout
		tcpClientHandler.SlaveId = byte(tcpConfig.WorkerID)
		tcpClientHandler.Logger = logger

		if err := tcpClientHandler.Connect(); err != nil {
			return nil, errors.Wrap(err, "failed to connect via TCP")
		}
		return &modbusClientHandler{handler: tcpClientHandler, closer: tcpClientHandler}, nil
	}

	if protocol.RTU != nil {
		var rtuConfig = protocol.RTU
		rtuClientHandler := modbus.NewRTUClientHandler(rtuConfig.Endpoint)
		rtuClientHandler.BaudRate = rtuConfig.BaudRate
		rtuClientHandler.DataBits = rtuConfig.DataBits
		rtuClientHandler.Parity = rtuConfig.Parity
		rtuClientHandler.StopBits = rtuConfig.StopBits
		rtuClientHandler.SlaveId = byte(rtuConfig.WorkerID)
		rtuClientHandler.Timeout = timeout
		rtuClientHandler.Logger = logger

		if err := rtuClientHandler.Connect(); err != nil {
			return nil, errors.Wrap(err, "failed to connect via RTU")
		}
		return &modbusClientHandler{handler: rtuClientHandler, closer: rtuClientHandler}, nil
	}

	return nil, errors.New("failed to create Modbus handler with empty protocol")
}
