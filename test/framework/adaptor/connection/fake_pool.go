package connection

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/registration"
)

func NewPool(ctx context.Context, receiver connection.Receiver) connection.Pool {
	ctx, cancel := context.WithCancel(ctx)

	return &pool{
		ctx:                ctx,
		cancel:             cancel,
		connectionReceiver: receiver,
	}
}

type pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	connectionReceiver connection.Receiver
}

func (p *pool) Start(n registration.EventNotifier) {
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

			n.NoticeHealthy()

			select {
			case <-doneC:
				return
			default:
			}
		}
	}()

	<-doneC
}

func (p *pool) Stop() error {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	return nil
}

func (p *pool) Create(link *v1alpha1.DeviceLink) error {
	return nil
}

func (p *pool) Delete(link *v1alpha1.DeviceLink) error {
	return nil
}

func (p *pool) SendData(link *v1alpha1.DeviceLink, device *unstructured.Unstructured) error {
	return nil
}
