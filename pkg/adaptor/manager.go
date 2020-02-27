package adaptor

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/listener"
	"github.com/rancher/octopus/pkg/adaptor/options"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/adaptor/validation"
)

type Manager interface {
	Start(<-chan struct{}) error
	GetPool() Pool
	RegisterRegistrationReceiver(receiver registration.Receiver)
	RegisterConnectionReceiver(receiver connection.Receiver)
}

func NewManager(opts options.Options) (Manager, error) {
	var socketPath = opts.SocketPath
	if socketPath == "" {
		socketPath = api.LimbSocket
	}
	if !filepath.IsAbs(socketPath) {
		return nil, errors.Errorf("unavailable socket path, must be an absolute path: %s", socketPath)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	var dir, file = filepath.Split(socketPath)

	return &manager{
		ctx:        ctx,
		cancel:     cancel,
		socketDir:  dir,
		socketName: file,
		pool:       NewPool(),
	}, nil
}

var log = ctrl.Log.WithName("adaptor").WithName("manager")

type manager struct {
	ctx    context.Context
	cancel context.CancelFunc

	socketDir  string
	socketName string
	server     *grpc.Server
	pool       Pool

	registrationReceiver registration.Receiver
	connectionReceiver   connection.Receiver
}

func (m *manager) Start(stopChan <-chan struct{}) error {
	defer m.stop()

	log.V(1).Info("starting")

	// TODO we also need a checkpoint like `kubelet/cm/devicemanager/manager.go#221`

	var socketPath = filepath.Join(m.socketDir, m.socketName)
	if err := os.MkdirAll(m.socketDir, 0755); err != nil {
		log.Error(err, "failed to create socket dir", "dir", socketPath)
	}

	// TODO validate SELinux

	// adaptor should connect this and use this event as a signal to re-register with Limb
	if err := m.cleanupStaleSockets(m.socketDir); err != nil {
		log.Error(err, "failed to clean up stable sockets under socket dir", "dir", socketPath)
	}

	var server, socket, err = listener.Listen(m.ctx, socketPath)
	if err != nil {
		log.Error(err, "failed to listen to socket while starting adaptor manager", "socket", socketPath)
		return err
	}
	api.RegisterRegistrationServiceServer(server, m)
	m.server = server

	var errChan = make(chan error)
	go func() {
		errChan <- server.Serve(socket)
	}()
	select {
	case err := <-errChan:
		return errors.Wrapf(err, "failed to serve the adaptor manager")
	case <-stopChan:
		return nil
	}
}

func (m *manager) GetPool() Pool {
	return m.pool
}

func (m *manager) RegisterRegistrationReceiver(receiver registration.Receiver) {
	m.registrationReceiver = receiver
}

func (m *manager) RegisterConnectionReceiver(receiver connection.Receiver) {
	m.connectionReceiver = receiver
}

func (m *manager) stop() {
	log.V(1).Info("stopping")

	// stops adaptor connection pool
	if m.pool != nil {
		m.pool.Stop()
		m.pool = nil
	}

	// stops grpc registration server
	if m.server != nil {
		m.server.Stop()
		m.server = nil
	}

	// cancels the context
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m *manager) addAdaptor(name, endpoint string) error {
	var connPool, err = connection.NewPool(m.ctx, filepath.Join(m.socketDir, endpoint), m.connectionReceiver)
	if err != nil {
		return errors.Wrap(err, "could not create adaptor connection pool")
	}

	// caches connection pool
	if connPoolStale := m.pool.Get(name); connPoolStale != nil {
		if err := connPoolStale.Stop(); err != nil {
			log.Error(err, "error while stopping stale adaptor connection pool")
		}
	}
	m.pool.Add(name, connPool)

	// starts connection pool
	var connNotifier = registration.NewEventNotifier(name, m.registrationReceiver)
	go connPool.Start(connNotifier)

	return nil
}

func (m *manager) cleanupStaleSockets(dir string) error {
	var socketDir, err = os.Open(dir)
	if err != nil {
		return err
	}
	defer socketDir.Close()

	contentNames, err := socketDir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range contentNames {
		if !validation.IsSocketFile(name) {
			continue
		}

		var filePath = filepath.Join(dir, name)
		var fileStat, err = os.Stat(filePath)
		if err != nil {
			log.Error(err, "failed to stat file", "file", filePath)
			continue
		}
		if fileStat.IsDir() {
			continue
		}

		if err = os.RemoveAll(filePath); err != nil {
			return err
		}
	}
	return nil
}
