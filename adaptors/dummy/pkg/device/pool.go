package device

import (
	"sync"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/object"
)

type Pool interface {
	Apply(device *v1alpha1.DummyDevice, parameters *Parameters, server api.AdaptorService_ConnectServer) error
}

type pool struct {
	sync.Mutex

	devices map[string]*dummyDevice
}

func NewPool() Pool {
	return &pool{
		devices: make(map[string]*dummyDevice),
	}
}

func (p *pool) Apply(device *v1alpha1.DummyDevice, parameters *Parameters, server api.AdaptorService_ConnectServer) error {
	p.Lock()
	defer p.Unlock()

	var deviceNamespaceName = object.GetNamespacedName(device)
	var deviceKey = deviceNamespaceName.String()
	var dummy = p.devices[deviceKey]

	if !device.Spec.On {
		dummy.stop()
		delete(p.devices, deviceKey)
		return nil
	}
	if dummy == nil {
		dummy = newDummy(deviceNamespaceName, server)
		p.devices[deviceKey] = dummy
		go dummy.start(parameters)
	}
	dummy.configure(device.Spec.Gear)

	return nil
}
