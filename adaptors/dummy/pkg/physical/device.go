package physical

import (
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
)

type Device interface {
	Configure(spec v1alpha1.DummyDeviceSpec)
	Shutdown()
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler) Device {
	return &device{
		log:     log,
		name:    name,
		handler: handler,
		ticker:  time.NewTicker(2 * time.Second),
	}
}

type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	ticker *time.Ticker
	status v1alpha1.DummyDeviceStatus
}

func (d *device) Configure(spec v1alpha1.DummyDeviceSpec) {
	if spec.On {
		d.on(spec.Gear)
	} else {
		d.off()
	}
}

func (d *device) Shutdown() {
	d.off()
	d.log.Info("Closed connection")
}

func (d *device) on(gear v1alpha1.DummyDeviceGear) {
	d.Lock()
	defer d.Unlock()

	if d.status.Gear == gear {
		return
	}

	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	// setup
	d.status.Gear = gear
	switch gear {
	case v1alpha1.Fast:
		d.status.RotatingSpeed = 200
	case v1alpha1.Middle:
		d.status.RotatingSpeed = 100
	case v1alpha1.Slow:
		d.status.RotatingSpeed = 0
	}
	d.handler(d.name, d.status)

	go d.mockPhysicalWatching(gear, d.stop)
}

func (d *device) off() {
	d.Lock()
	defer d.Unlock()

	if d.stop != nil {
		close(d.stop)
	}
	d.status.Gear = ""
	d.status.RotatingSpeed = 0

	d.handler(d.name, d.status)
}

func (d *device) mockPhysicalWatching(gear v1alpha1.DummyDeviceGear, stop <-chan struct{}) {
	d.log.Info("Mocking started")
	defer func() {
		d.log.Info("Mocking finished")
	}()

	var duration time.Duration
	switch gear {
	case v1alpha1.Slow:
		duration = 3 * time.Second
	case v1alpha1.Middle:
		duration = 2 * time.Second
	case v1alpha1.Fast:
		duration = 1 * time.Second
	}
	var ticker = time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		d.Lock()

		func() {
			defer d.Unlock()

			switch d.status.Gear {
			case v1alpha1.Fast:
				if d.status.RotatingSpeed < 300 {
					d.status.RotatingSpeed++
				}
			case v1alpha1.Middle:
				if d.status.RotatingSpeed < 200 {
					d.status.RotatingSpeed++
				}
			case v1alpha1.Slow:
				if d.status.RotatingSpeed < 100 {
					d.status.RotatingSpeed++
				}
			}
			d.handler(d.name, d.status)
		}()

		select {
		case <-stop:
			return
		default:
		}
	}
}
