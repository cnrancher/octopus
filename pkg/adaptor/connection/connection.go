package connection

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

// Receiver is used to receive the changes from adaptor
type Receiver interface {
	ReceiveDeviceChanges(adaptorName string, deviceObserved *unstructured.Unstructured, err error)
}

func newConnection(ctx context.Context, clientConn *grpc.ClientConn) (*connection, error) {
	var connClient, err = api.NewAdaptorServiceClient(clientConn).Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &connection{
		client: connClient,
	}, nil
}

type connection struct {
	client api.AdaptorService_ConnectClient
}

func (c *connection) receive(adaptorName string, receiver Receiver) {
	if receiver == nil {
		return
	}

	for {
		var resp, err = c.client.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			receiver.ReceiveDeviceChanges(adaptorName, nil, errors.Wrap(err, "failed on receiving"))
			return
		}

		var device = resp.GetDevice()
		var deviceObserved = &unstructured.Unstructured{Object: make(map[string]interface{})}
		if err := deviceObserved.UnmarshalJSON(device); err != nil {
			receiver.ReceiveDeviceChanges(adaptorName, nil, errors.Wrap(err, "failed on unmarshal"))
			continue
		}

		receiver.ReceiveDeviceChanges(adaptorName, deviceObserved, nil)
	}
}

func (c *connection) send(adaptorParameters *runtime.RawExtension, deviceDesired *unstructured.Unstructured) error {
	var device, err = deviceDesired.DeepCopy().MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshal device into JSON")
	}
	var parameters = adaptorParameters.DeepCopy().Raw

	return c.client.Send(&api.ConnectionRequest{
		Parameters: parameters,
		Device:     device,
	})
}

func (c *connection) close() error {
	return c.client.CloseSend()
}
