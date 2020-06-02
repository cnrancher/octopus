package adaptor

import (
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("suctioncup").WithName("adaptors")

type Adaptors struct {
	index *sync.Map
}

func NewAdaptors() Adaptors {
	return Adaptors{
		index: &sync.Map{},
	}
}

func (c Adaptors) Get(nameOrEndpoint string) Adaptor {
	if aa, exist := c.index.Load(nameOrEndpoint); exist {
		return aa.(Adaptor)
	}
	return nil
}

func (c Adaptors) Delete(nameOrEndpoint string) {
	if aa, exist := c.index.Load(nameOrEndpoint); exist {
		var adaptor = aa.(Adaptor)
		if err := adaptor.Stop(); err != nil {
			log.Error(err, "Failed to stop adaptor", "adaptor", adaptor.GetName())
		}
		c.index.Delete(adaptor.GetEndpoint())
		c.index.Delete(adaptor.GetName())
	}
}

func (c Adaptors) Put(adaptor Adaptor) {
	if aa, exist := c.index.LoadOrStore(adaptor.GetName(), adaptor); exist {
		var staleAdaptor = aa.(Adaptor)
		if err := staleAdaptor.Stop(); err != nil {
			log.Error(err, "Failed to stop stale adaptor", "adaptor", staleAdaptor.GetName())
		}
		c.index.Store(adaptor.GetName(), adaptor)
	}
	c.index.Store(adaptor.GetEndpoint(), adaptor)
}

func (c Adaptors) Cleanup() {
	c.index.Range(func(nameOrEndpoint, aa interface{}) bool {
		// delete
		c.index.Delete(nameOrEndpoint)
		// close adaptor
		var adaptor = aa.(Adaptor)
		if err := adaptor.Stop(); err != nil {
			log.Error(err, "Failed to stop adaptor", "adaptor", adaptor.GetName())
		}

		return true
	})
}
