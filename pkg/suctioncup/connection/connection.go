package connection

import (
	"context"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/event"
)

type Connection interface {
	// GetAdaptorName returns the name of adaptor
	GetAdaptorName() string

	// GetName returns the name of connection
	GetName() types.NamespacedName

	// Send sends the parameters and desired data to connection
	Send(parameters, desired []byte) error

	// Stop stops the connection
	Stop() error
}

func NewConnection(adaptorName string, name types.NamespacedName, clientConn *grpc.ClientConn, notifier event.ConnectionNotifier) (Connection, error) {
	var conn, err = api.NewConnectionClient(clientConn).Connect(context.Background())
	if err != nil {
		return nil, err
	}

	var c = &connection{
		adaptorName: adaptorName,
		name:        name,
		conn:        conn,
		notifier:    notifier,
	}
	go c.receive()

	return c, nil
}

type connection struct {
	adaptorName string
	name        types.NamespacedName
	conn        api.Connection_ConnectClient
	notifier    event.ConnectionNotifier
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

func (c *connection) Send(parameters, device []byte) error {
	return c.conn.Send(&api.ConnectRequest{
		Parameters: parameters,
		Device:     device,
	})
}

func (c *connection) stop() error {
	return c.conn.CloseSend()
}

func (c *connection) receive() {
	defer c.stop()

	for {
		var resp, err = c.conn.Recv()
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
		} else {
			c.notifier.NoticeConnectionReceivedData(
				c.adaptorName,
				c.name,
				resp.GetDevice(),
			)
		}
	}
}
