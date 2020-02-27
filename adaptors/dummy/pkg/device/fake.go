package device

import (
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

type dummyDevice struct {
	sync.Mutex
	v1alpha1.DummyDeviceStatus

	isStop bool
	name   types.NamespacedName
	server api.AdaptorService_ConnectServer
}

func newDummy(name types.NamespacedName, server api.AdaptorService_ConnectServer) *dummyDevice {
	return &dummyDevice{
		name:   name,
		server: server,
	}
}

func (d *dummyDevice) start(_ *Parameters) {
	for _ = range time.Tick(2 * time.Second) {
		d.Lock()
		if d.isStop {
			return
		}

		switch d.Gear {
		case v1alpha1.Fast:
			if d.RotatingSpeed < 300 {
				d.RotatingSpeed += 1
			}
		case v1alpha1.Middle:
			if d.RotatingSpeed < 200 {
				d.RotatingSpeed += 1
			}
		case v1alpha1.Slow:
			if d.RotatingSpeed < 100 {
				d.RotatingSpeed += 1
			}
		}

		var device = v1alpha1.DummyDevice{
			TypeMeta: v1.TypeMeta{
				Kind:       "DummyDevice",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Namespace: d.name.Namespace,
				Name:      d.name.Name,
			},
			Status: d.DummyDeviceStatus,
		}
		var b, _ = jsoniter.Marshal(device)
		_ = d.server.Send(&api.ConnectionResponse{Device: b})

		d.Unlock()
	}

}

func (d *dummyDevice) stop() {
	d.Lock()
	defer d.Unlock()

	var device = v1alpha1.DummyDevice{
		TypeMeta: v1.TypeMeta{
			Kind:       "DummyDevice",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: d.name.Namespace,
			Name:      d.name.Name,
		},
	}
	var b, _ = jsoniter.Marshal(device)

	d.isStop = true
	_ = d.server.Send(&api.ConnectionResponse{Device: b})
}

func (d *dummyDevice) configure(gear v1alpha1.DummyDeviceGear) {
	d.Lock()
	defer d.Unlock()

	if d.Gear == gear {
		return
	}

	d.Gear = gear
	switch gear {
	case v1alpha1.Fast:
		d.RotatingSpeed = 200
	case v1alpha1.Middle:
		d.RotatingSpeed = 100
	case v1alpha1.Slow:
		d.RotatingSpeed = 0
	}
}
