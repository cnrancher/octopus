package adaptor

import (
	"sync"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
)

type Pool interface {
	// Exist judges whether the adaptor that matches name exist.
	Exist(adaptorName string) bool
	// Get returns the connection pool of the adaptor that matches the name.
	Get(adaptorName string) connection.Pool
	// Add adds the connection pool that related the name.
	Add(adaptorName string, connPool connection.Pool)
	// Delete deletes the connection pool that related the name.
	Delete(adaptorName string)
	// Stop stops the whole pool.
	Stop()
	// CreateConnection is a shortcut for creating the connection of a link.
	CreateConnection(link *edgev1alpha1.DeviceLink) error
	// DeleteConnection is a shortcut for deleting the connection of a link.
	DeleteConnection(link *edgev1alpha1.DeviceLink) error
	// SendDataToConnection is a shortcut for sending the device information to the connection of a link.
	SendDataToConnection(link *edgev1alpha1.DeviceLink, device *unstructured.Unstructured) error
}

func NewPool() Pool {
	return &pools{
		conns: make(map[string]connection.Pool),
	}
}

type pools struct {
	sync.RWMutex

	conns map[string]connection.Pool
}

func (p *pools) Exist(adaptorName string) bool {
	return p.Get(adaptorName) != nil
}

func (p *pools) Get(adaptorName string) connection.Pool {
	p.RLock()
	defer p.RUnlock()

	return p.conns[adaptorName]
}

func (p *pools) Add(adaptorName string, connPool connection.Pool) {
	p.Lock()
	defer p.Unlock()

	p.conns[adaptorName] = connPool
}

func (p *pools) Delete(adaptorName string) {
	p.Lock()
	defer p.Unlock()

	delete(p.conns, adaptorName)
}

func (p *pools) Stop() {
	p.Lock()
	defer p.Unlock()

	for _, pool := range p.conns {
		_ = pool.Stop()
	}
}

func (p *pools) CreateConnection(link *edgev1alpha1.DeviceLink) error {
	var pool = p.Get(link.Status.Adaptor.Name)
	if pool == nil {
		return errors.New("adaptor isn't existed")
	}
	return pool.Create(link)
}

func (p *pools) DeleteConnection(link *edgev1alpha1.DeviceLink) error {
	var pool = p.Get(link.Status.Adaptor.Name)
	if pool == nil {
		return errors.New("adaptor isn't existed")
	}
	return pool.Delete(link)
}

func (p *pools) SendDataToConnection(link *edgev1alpha1.DeviceLink, device *unstructured.Unstructured) error {
	var pool = p.Get(link.Status.Adaptor.Name)
	if pool == nil {
		return errors.New("adaptor isn't existed")
	}
	return pool.SendData(link, device)
}
