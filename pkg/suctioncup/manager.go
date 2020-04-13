package suctioncup

import (
	"runtime"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/adaptor"
	"github.com/rancher/octopus/pkg/suctioncup/event"
	"github.com/rancher/octopus/pkg/suctioncup/registration"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

var log = ctrl.Log.WithName("suctioncup").WithName("manager")

func NewManager() (Manager, error) {
	var adaptors = adaptor.NewAdaptors()
	var queue = event.NewQueue("adaptor.manager")
	return NewManagerWith(adaptors, queue)
}

func NewManagerWith(adaptors adaptor.Adaptors, queue event.Queue) (Manager, error) {
	var regSrv, err = registration.NewServer(api.LimbSocket, adaptors, queue)
	if err != nil {
		return nil, err
	}

	return &manager{
		adaptors:           adaptors,
		queue:              queue,
		registrationServer: regSrv,
	}, nil
}

type manager struct {
	adaptors           adaptor.Adaptors
	queue              event.Queue
	registrationServer registration.Server
}

func (m *manager) RegisterAdaptorHandler(handler event.AdaptorHandler) {
	m.queue.RegisterAdaptorHandler(handler)
}

func (m *manager) RegisterConnectionHandler(handler event.ConnectionHandler) {
	m.queue.RegisterConnectionHandler(handler)
}

func (m *manager) GetNeurons() Neurons {
	return m
}

func (m *manager) Start(stop <-chan struct{}) error {
	defer utilruntime.HandleCrash(handler.NewPanicsLogHandler(log))
	defer m.stop()

	// starts reconcile
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(index int) {
			log.Info("Starting reconciling queue", "index", index)
			wait.Until(m.queue.Reconcile, 1*time.Second, stop)
		}(i)
	}

	// serves adaptor registration server
	log.Info("Starting registration server")
	var err = m.registrationServer.Start(stop)
	return err
}

func (m *manager) stop() {
	m.adaptors.Cleanup()
	m.queue.ShutDown()
}
