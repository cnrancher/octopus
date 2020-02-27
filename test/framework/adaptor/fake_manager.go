package adaptor

import (
	"context"

	"github.com/rancher/octopus/pkg/adaptor"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	fakeconnection "github.com/rancher/octopus/test/framework/adaptor/connection"
)

func NewManager() *Manager {
	var ctx, cancel = context.WithCancel(context.Background())

	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		pool:   adaptor.NewPool(),
	}
}

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc

	pool adaptor.Pool

	registrationReceiver registration.Receiver
	connectionReceiver   connection.Receiver
}

func (m *Manager) Start(stopChan <-chan struct{}) error {
	defer m.stop()

	<-stopChan
	return nil
}

func (m *Manager) stop() {
	// stops adaptor connection pool
	if m.pool != nil {
		m.pool.Stop()
		m.pool = nil
	}

	// cancels the context
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m *Manager) GetPool() adaptor.Pool {
	return m.pool
}

func (m *Manager) RegisterRegistrationReceiver(receiver registration.Receiver) {
	m.registrationReceiver = receiver
}

func (m *Manager) RegisterConnectionReceiver(receiver connection.Receiver) {
	m.connectionReceiver = receiver
}

// AddAdaptor simulates to add an adaptor
func (m *Manager) AddAdaptor(name string) {
	var connPool = fakeconnection.NewPool(m.ctx, m.connectionReceiver)

	// caches connection pool
	if connPoolStale := m.pool.Get(name); connPoolStale != nil {
		_ = connPoolStale.Stop()
	}
	m.pool.Add(name, connPool)

	// starts connection pool
	var connNotifier = registration.NewEventNotifier(name, m.registrationReceiver)
	go connPool.Start(connNotifier)
}

// DeleteAdaptor simulates to delete an adaptor
func (m *Manager) DeleteAdaptor(name string) {
	if m.pool.Exist(name) {
		_ = m.pool.Get(name).Stop()
		m.pool.Delete(name)
	}
}
