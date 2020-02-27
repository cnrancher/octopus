package connection

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/dialer"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/util/object"
)

type Pool interface {
	// Start starts the connection pool of the adaptor,
	// any subscriber could be notified by the notifier during this phase.
	Start(notifier registration.EventNotifier)
	// Stop stops the connection pool.
	Stop() error
	// Create creates the connection of a link.
	Create(link *edgev1alpha1.DeviceLink) error
	// Delete deletes the connection of a link.
	Delete(link *edgev1alpha1.DeviceLink) error
	// SendDataToConnection sends the device information to the connection of a link.
	SendData(link *edgev1alpha1.DeviceLink, device *unstructured.Unstructured) error
}

func NewPool(ctx context.Context, socketPath string, receiver Receiver) (Pool, error) {
	var clientConn, err = dialer.Dial(ctx, socketPath)
	if err != nil {
		return nil, err
	}

	keepaliveClient, err := api.NewAdaptorServiceClient(clientConn).KeepAlive(ctx)
	if err != nil {
		return nil, err
	}
	if err = keepaliveClient.Send(&api.Void{}); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &pool{
		ctx:                ctx,
		cancel:             cancel,
		clientConn:         clientConn,
		keepaliveClient:    keepaliveClient,
		connectionReceiver: receiver,
		connections:        make(map[string]*connection),
	}, nil
}

type pool struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	clientConn      *grpc.ClientConn
	keepaliveClient api.AdaptorService_KeepAliveClient

	connectionReceiver Receiver
	connections        map[string]*connection
}

func (p *pool) Start(n registration.EventNotifier) {
	defer p.clientConn.Close()
	defer p.keepaliveClient.CloseSend()
	defer n.NoticeStopped()

	n.NoticeStarted()

	var doneC = p.ctx.Done()
	var ticker = time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	go func() {
		for _ = range ticker.C {
			select {
			case <-doneC:
				return
			default:
			}

			if err := p.keepaliveClient.Send(&api.Void{}); err != nil {
				if err == io.EOF {
					return
				}
				n.NoticeUnhealthy(err.Error())
			} else {
				n.NoticeHealthy()
			}

			select {
			case <-doneC:
				return
			default:
			}
		}
	}()

	<-doneC
}

func (p *pool) Create(link *edgev1alpha1.DeviceLink) error {
	if link == nil {
		return errors.New("link is nil")
	}

	p.RLock()
	defer p.RUnlock()

	var connName = object.GetNamespacedName(link).String()
	if _, exist := p.connections[connName]; exist {
		return nil
	}

	var conn, err = newConnection(p.ctx, p.clientConn)
	if err != nil {
		return err
	}
	go conn.receive(link.Spec.Adaptor.Name, p.connectionReceiver)

	p.connections[connName] = conn
	return nil
}

func (p *pool) Delete(link *edgev1alpha1.DeviceLink) error {
	if link == nil {
		return errors.New("link is nil")
	}

	p.Lock()
	defer p.Unlock()

	var connName = object.GetNamespacedName(link).String()
	if conn, exist := p.connections[connName]; exist {
		if err := conn.close(); err != nil {
			return err
		}
		delete(p.connections, connName)
	}
	// ignores adaptor is not exited
	return nil
}

func (p *pool) SendData(link *edgev1alpha1.DeviceLink, device *unstructured.Unstructured) error {
	if link == nil {
		return errors.New("link is nil")
	}
	if device == nil {
		return errors.New("device is nil")
	}

	p.RLock()
	defer p.RUnlock()

	var connName = object.GetNamespacedName(link).String()
	if conn, exist := p.connections[connName]; exist {
		return conn.send(link.Spec.Adaptor.Parameters, device)
	}
	return errors.New("adaptor is not existed")
}

func (p *pool) Stop() error {
	p.Lock()
	defer p.Unlock()

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	return nil
}
