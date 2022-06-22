package physical

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/ble/pkg/metadata"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
)

var (
	ErrGATTCharacteristicIncompatibleRead   = errors.Errorf("target GATT characteristic is not readable")
	ErrGATTCharacteristicIncompatibleWrite  = errors.Errorf("target GATT characteristic is not writable")
	ErrGATTCharacteristicIncompatibleNotify = errors.Errorf("target GATT characteristic is not notifiable")
	ErrGATTPeripheralConnectionClosed       = errors.New("target GATT peripheral is terminated the connection")
)

// BluetoothPeripheralConnectionLostHandler specifies to handle the connection lost of Bluetooth peripheral.
type BluetoothPeripheralConnectionLostHandler func(ble.Client, error)

// BluetoothPeripheralConnectionOptions specifies the options for connecting Bluetooth peripheral.
type BluetoothPeripheralConnectionOptions struct {
	AutoReconnect                  bool
	MaxReconnectInterval           time.Duration
	ConnectMTU                     int
	ConnectTimeout                 time.Duration
	OnlySubscribeNotificationValue bool
	OnlyWriteValueWithoutResponse  bool
	OnConnectionLost               BluetoothPeripheralConnectionLostHandler
}

type operationType int

const (
	read operationType = iota
	write
	subscribe
	clearSubscriptions
	reconnect
)

type operationPayload struct {
	operation      operationType
	service        string
	characteristic string
	data           []byte
	err            error
	notifyHandler  ble.NotificationHandler
}

type subscription struct {
	handler        ble.NotificationHandler
	service        string
	characteristic string
}

// BluetoothPeripheral represents a Bluetooth peripheral.
type BluetoothPeripheral struct {
	sync.Once

	log      logr.Logger
	done     chan struct{}
	request  chan operationPayload
	response chan operationPayload
	addr     ble.Addr

	options   BluetoothPeripheralConnectionOptions
	connected atomic.Bool
	cli       ble.Client
	subs      map[string]subscription
}

// ReadCharacteristic reads the GATT characteristic data via given {service, characteristic} UUID.
func (p *BluetoothPeripheral) ReadCharacteristic(service, characteristic string) ([]byte, error) {
	var req = operationPayload{
		operation:      read,
		service:        service,
		characteristic: characteristic,
	}
	p.request <- req

	var resp = <-p.response
	p.log.V(4).Info("Read characteristic", "service", service, "characteristic", characteristic)
	return resp.data, resp.err
}

// WriteCharacteristic writes data to the GATT characteristic matched given {service, characteristic} UUID.
func (p *BluetoothPeripheral) WriteCharacteristic(service, characteristic string, data []byte) error {
	var req = operationPayload{
		operation:      write,
		service:        service,
		characteristic: characteristic,
		data:           data,
	}
	p.request <- req

	var resp = <-p.response
	p.log.V(4).Info("Write characteristic", "service", service, "characteristic", characteristic)
	return resp.err
}

// SubscribeCharacteristic subscribes the GATT characteristic matched given {service, characteristic} UUID.
func (p *BluetoothPeripheral) SubscribeCharacteristic(service, characteristic string, handler ble.NotificationHandler) error {
	var req = operationPayload{
		operation:      subscribe,
		service:        service,
		characteristic: characteristic,
		notifyHandler:  handler,
	}
	p.request <- req

	var resp = <-p.response
	p.log.V(4).Info("Subscribe characteristic", "service", service, "characteristic", characteristic)
	return resp.err
}

// ClearSubscriptions clears all GATT characteristic subscriptions.
func (p *BluetoothPeripheral) ClearSubscriptions() error {
	var req = operationPayload{operation: clearSubscriptions}
	p.request <- req

	var resp = <-p.response
	p.log.V(4).Info("Clear all subscriptions")
	return resp.err
}

func (p *BluetoothPeripheral) Close() error {
	if p.done == nil {
		return nil
	}
	close(p.done)
	p.done = nil

	var err error
	if p.cli != nil {
		err = p.cli.CancelConnection()
	}
	p.log.V(4).Info("Closed")
	return err
}

// SetConnectionOptions configures the connection options.
func (p *BluetoothPeripheral) SetConnectionOptions(options BluetoothPeripheralConnectionOptions) *BluetoothPeripheral {
	p.options = options
	return p
}

func (p *BluetoothPeripheral) Start() {
	p.Do(func() {
		go p.loop()
	})
}

// loop starts a loop to receive the operation.
func (p *BluetoothPeripheral) loop() {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	if p.done != nil {
		return
	}
	p.done = make(chan struct{})

	p.log.Info("Start receiving operations")
	defer func() {
		p.log.Info("Finished receiving operations")
	}()

	// NB(thxCode) in order to reduce the dependence on the lock,
	// chan is used here to solve the reconnection after the connection is broken.
	for {
		var req operationPayload
		select {
		case <-p.done:
			return
		case req = <-p.request:
		}

		var cli ble.Client
		var cliDialedErr = func() error {
			if p.connected.Load() {
				cli = p.cli
			} else {
				var cerr error
				cli, cerr = p.connect()
				if cerr != nil {
					return cerr
				}
				p.cli = cli
				p.connected.Store(true)

				go func() {
					select {
					case <-p.done:
						return
					case <-cli.Disconnected():
						p.log.V(5).Info("the GATT peripheral is disconnected")
					}

					p.connected.Store(false)
					if p.options.OnConnectionLost != nil {
						go p.options.OnConnectionLost(p.cli, ErrGATTPeripheralConnectionClosed)
					}

					if !p.options.AutoReconnect {
						return
					}
					p.log.V(5).Info("Start to reconnect the GATT peripheral")

					var reconnectSleep = time.Second
					for {
						var req = operationPayload{
							operation: reconnect,
						}
						p.request <- req

						var resp = <-p.response
						if resp.err == nil {
							break
						}

						p.log.Error(resp.err, fmt.Sprintf("Failed to reconnect the GATT peripheral sleeping for %s", reconnectSleep))
						select {
						case <-p.done:
							return
						case <-time.After(reconnectSleep):
						}
						if reconnectSleep < p.options.MaxReconnectInterval {
							reconnectSleep *= 2
						}
						if reconnectSleep > p.options.MaxReconnectInterval {
							reconnectSleep = p.options.MaxReconnectInterval
						}
					}
				}()
			}
			return nil
		}()
		if cliDialedErr != nil {
			p.response <- operationPayload{
				operation: req.operation,
				err:       cliDialedErr,
			}
			continue
		}

		switch req.operation {
		case read:
			var char, err = findCharacteristic(cli.Profile(), req.service, req.characteristic)
			if err != nil {
				p.response <- operationPayload{
					operation: req.operation,
					err:       err,
				}
				break
			}
			p.log.V(5).Info("READ: found GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			if char.Property&(ble.CharRead) == 0 {
				p.response <- operationPayload{
					operation: req.operation,
					err:       ErrGATTCharacteristicIncompatibleRead,
				}
			}
			p.log.V(5).Info("READ: validated GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			data, err := cli.ReadLongCharacteristic(char)
			if err != nil {
				p.log.V(5).Info("READ: read GATT characteristic", "service", req.service, "characteristic", req.characteristic)
			}
			p.response <- operationPayload{
				operation: req.operation,
				data:      data,
				err:       err,
			}
		case write:
			var char, err = findCharacteristic(cli.Profile(), req.service, req.characteristic)
			if err != nil {
				p.response <- operationPayload{
					operation: req.operation,
					err:       err,
				}
				break
			}
			p.log.V(5).Info("WRITE: found GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			if char.Property&(ble.CharWriteNR|ble.CharWrite) == 0 {
				p.response <- operationPayload{
					operation: req.operation,
					err:       ErrGATTCharacteristicIncompatibleWrite,
				}
				break
			}
			p.log.V(5).Info("WRITE: validated GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			if p.options.OnlyWriteValueWithoutResponse && (char.Property&ble.CharWrite&ble.CharWriteNR != 0) {
				err = cli.WriteCharacteristic(char, req.data, true)
			} else {
				err = cli.WriteCharacteristic(char, req.data, char.Property&ble.CharWriteNR != 0)
			}
			if err != nil {
				p.log.V(5).Info("WRITE: wrote GATT characteristic", "service", req.service, "characteristic", req.characteristic)
			}
			p.response <- operationPayload{
				operation: req.operation,
				err:       err,
			}
		case subscribe:
			var char, err = findCharacteristic(cli.Profile(), req.service, req.characteristic)
			if err != nil {
				p.response <- operationPayload{
					operation: req.operation,
					err:       err,
				}
				break
			}
			p.log.V(5).Info("SUBSCRIBE: found GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			if char.Property&(ble.CharNotify|ble.CharIndicate) == 0 {
				p.response <- operationPayload{
					operation: req.operation,
					err:       ErrGATTCharacteristicIncompatibleNotify,
				}
				break
			}
			p.log.V(5).Info("SUBSCRIBE: validated GATT characteristic", "service", req.service, "characteristic", req.characteristic)

			// always record notification
			p.subs[fmt.Sprintf("%s/%s", req.service, req.characteristic)] = subscription{
				handler:        req.notifyHandler,
				service:        req.service,
				characteristic: req.characteristic,
			}

			if p.options.OnlySubscribeNotificationValue && (char.Property&ble.CharNotify&ble.CharIndicate != 0) {
				err = cli.Subscribe(char, false, req.notifyHandler)
			} else {
				err = cli.Subscribe(char, char.Property&ble.CharIndicate != 0, req.notifyHandler)
			}
			if err != nil {
				p.log.V(5).Info("SUBSCRIBE: subscribed GATT characteristic", "service", req.service, "characteristic", req.characteristic)
			}
			p.response <- operationPayload{
				operation: req.operation,
				err:       err,
			}
		case clearSubscriptions:
			var err = cli.ClearSubscriptions()
			if err != nil {
				p.log.V(5).Info("CLEAR_SUBSCRIPTIONS: cleared GATT subscriptions", "service", req.service, "characteristic", req.characteristic)
			}
			p.response <- operationPayload{
				operation: req.operation,
				err:       err,
			}
		case reconnect:
			p.log.V(5).Info("RECONNECT")
			fallthrough
		default:
			p.response <- operationPayload{
				operation: req.operation,
			}
		}
	}
}

// connect connects to the peripheral.
func (p *BluetoothPeripheral) connect() (ble.Client, error) {
	var addr = p.addr
	var options = p.options

	// dials
	var timeout = options.ConnectTimeout
	p.log.V(5).Info("Dialing GATT peripheral", "timeout", timeout)
	var timeoutCtx, cancelTimeoutCtx = context.WithTimeout(ctx, timeout)
	defer cancelTimeoutCtx()
	cli, err := central.Dial(timeoutCtx, addr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect GATT peripheral %s", addr)
	}

	// discovers
	p.log.V(5).Info("Discovering GATT peripheral")
	_, err = cli.DiscoverProfile(true)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to discover GATT peripheral %s", addr)
	}

	// exchanges MTU
	if options.ConnectMTU > 0 {
		p.log.V(5).Info("Exchanging MTU for GATT peripheral")
		mtu, err := cli.ExchangeMTU(options.ConnectMTU)
		if err != nil {
			return nil, errors.Wrap(err, "failed to exchange expected MTU with GATT peripheral")
		}
		if mtu < options.ConnectMTU {
			log.V(5).Info("GATT peripheral gets a lower MTU", "expected", options.ConnectMTU, "actual", mtu)
		}
	}

	// re-subscribes notification
	for k, sub := range p.subs {
		p.log.V(5).Info("Resubscribing GATT peripheral", "service", sub.service, "characteristic", sub.characteristic)
		var char, err = findCharacteristic(cli.Profile(), sub.service, sub.characteristic)
		if err != nil {
			return nil, err
		}
		err = cli.Subscribe(char, char.Property&ble.CharIndicate != 0, sub.handler)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to re-subscribe characteristic (%s) of GATT peripheral %s", k, addr)
		}
	}

	return cli, nil
}

func NewBluetoothPeripheral(adv ble.Advertisement) *BluetoothPeripheral {
	var logger = log.WithName("ble.client").WithValues("endpoint", fmt.Sprintf("%s(%s)", adv.Addr(), adv.LocalName()))
	var peripheral = &BluetoothPeripheral{
		log:      logger,
		request:  make(chan operationPayload),
		response: make(chan operationPayload),
		addr:     adv.Addr(),
		subs:     make(map[string]subscription),
	}
	return peripheral
}

// findCharacteristic searches the GATT characteristic from given peripheral profile with given {service, characteristic} UUID.
func findCharacteristic(profile *ble.Profile, service, characteristic string) (*ble.Characteristic, error) {
	if profile == nil {
		return nil, errors.New("the profile of GATT peripheral is not found")
	}

	var characteristicUUID, err = ble.Parse(characteristic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse GATT characteristic UUID")
	}

	var char *ble.Characteristic
	if service != "" {
		var serviceUUID, err = ble.Parse(service)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse GATT service UUID")
		}

		var svc = profile.FindService(ble.NewService(serviceUUID))
		if svc == nil {
			return nil, errors.Errorf("failed to find GATT service %s", service)
		}

		for _, c := range svc.Characteristics {
			if c.UUID.Equal(characteristicUUID) {
				char = c
				break
			}
		}
	} else {
		char = profile.FindCharacteristic(ble.NewCharacteristic(characteristicUUID))
	}

	if char == nil {
		if service != "" {
			return nil, errors.Errorf("failed to find GATT characteristic %s belong to service %s", characteristic, service)
		}
		return nil, errors.Errorf("failed to find GATT characteristic %s", characteristic)
	}
	return char, nil
}
