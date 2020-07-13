package connection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/event"
)

type Connection interface {
	// GetAdaptorName returns the name of adaptor
	GetAdaptorName() string

	// GetName returns the name of connection
	GetName() types.NamespacedName

	// Send sends the device model, desired data and references to connection
	Send(model *metav1.TypeMeta, device []byte, references map[string]*api.ConnectRequestReferenceEntry) error

	// Stop stops the connection
	Stop() error

	// IsStop returns true if the connection is stopped
	IsStop() bool
}

func NewConnection(adaptorName string, name types.NamespacedName, clientConn *grpc.ClientConn, notifier event.ConnectionNotifier) (Connection, error) {
	var conn, err = api.NewConnectionClient(clientConn).Connect(context.Background())
	if err != nil {
		return nil, err
	}

	var c = &connection{
		adaptorName:     adaptorName,
		name:            name,
		conn:            conn,
		notifier:        notifier,
		interruptSignal: make(chan struct{}),
		interruptError:  make(chan error),
	}
	go c.receive()
	return c, nil
}

type connection struct {
	stopped     atomic.Bool
	adaptorName string
	name        types.NamespacedName
	conn        api.Connection_ConnectClient
	notifier    event.ConnectionNotifier

	interruptSignal chan struct{}
	interruptError  chan error
}

func (c *connection) GetAdaptorName() string {
	return c.adaptorName
}

func (c *connection) GetName() types.NamespacedName {
	return c.name
}

func (c *connection) Stop() error {
	return c.stop()
}

func (c *connection) IsStop() bool {
	return c.stopped.Load()
}

func (c *connection) Send(model *metav1.TypeMeta, device []byte, references map[string]*api.ConnectRequestReferenceEntry) (err error) {
	defer func() {
		if err != nil {
			_ = c.stop()
		}
	}()

	go func() {
		defer func() {
			if recover() != nil {
				// nothing to do
			}
		}()
		c.interruptSignal <- struct{}{}
	}()
	if err = c.conn.Send(&api.ConnectRequest{
		Model:      model,
		Device:     device,
		References: references,
	}); err != nil {
		return
	}

	// TODO should we parameterize the sending timeout?
	var timeoutDuration = 90 * time.Second
	var timeout = time.NewTimer(timeoutDuration)
	defer timeout.Stop()
	select {
	case err = <-c.interruptError:
		return
	case <-timeout.C:
		return fmt.Errorf("timeout to send data in %v", timeoutDuration)
	}
}

func (c *connection) stop() error {
	var err error
	if c.stopped.CAS(false, true) {
		err = c.conn.CloseSend()
		close(c.interruptSignal)
		close(c.interruptError)
	}
	return err
}

func (c *connection) receive() {
	defer c.stop()

	for {
		var resp, err = c.conn.Recv()
		select {
		case _, active := <-c.interruptSignal:
			if !active {
				return
			}

			if err != nil {
				c.interruptError <- err
				return
			}
			if resp == nil {
				c.interruptError <- errors.New("failed to receive data")
				return
			}

			if resp.GetErrorMessage() != "" {
				c.interruptError <- errors.New(resp.GetErrorMessage())
			} else {
				c.notifier.NoticeConnectionReceivedData(
					c.adaptorName,
					c.name,
					resp.GetDevice(),
				)
				c.interruptError <- nil
			}
			continue
		default:
		}

		if err != nil {
			// NB(thxCode) active shutdown means that the Stop() has been called
			if isActiveClosed(err) {
				return
			}

			// NB(thxCode) passive shutdown means that the server has been closed
			if isPassiveClosed(err) {
				c.notifier.NoticeConnectionClosed(
					c.adaptorName,
					c.name,
				)
				return
			}

			c.notifier.NoticeConnectionReceivedError(
				c.adaptorName,
				c.name,
				err,
			)
			return
		}

		if resp == nil {
			return
		}

		if resp.GetErrorMessage() != "" {
			c.notifier.NoticeConnectionReceivedError(
				c.adaptorName,
				c.name,
				errors.New(resp.GetErrorMessage()),
			)
		} else {
			c.notifier.NoticeConnectionReceivedData(
				c.adaptorName,
				c.name,
				resp.GetDevice(),
			)
		}
	}
}
