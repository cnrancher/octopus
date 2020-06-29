package connection

import (
	"context"

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
		c.interruptSignal <- struct{}{}
	}()
	if err = c.conn.Send(&api.ConnectRequest{
		Model:      model,
		Device:     device,
		References: references,
	}); err == nil {
		err = <-c.interruptError
	}

	return err
}

func (c *connection) stop() error {
	if c.stopped.CAS(false, true) {
		close(c.interruptSignal)
		close(c.interruptError)
		return c.conn.CloseSend()
	}
	return nil
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

			c.notifier.NoticeConnectionReceivedData(
				c.adaptorName,
				c.name,
				resp.GetDevice(),
			)
			c.interruptError <- nil
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

		c.notifier.NoticeConnectionReceivedData(
			c.adaptorName,
			c.name,
			resp.GetDevice(),
		)
	}
}
