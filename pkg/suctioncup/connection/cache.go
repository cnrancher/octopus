package connection

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("suctioncup").WithName("connections")

type Connections struct {
	index *sync.Map
}

func NewConnections() Connections {
	return Connections{
		index: &sync.Map{},
	}
}

func (c Connections) Get(name types.NamespacedName) Connection {
	if cc, exist := c.index.Load(name); exist {
		return cc.(Connection)
	}
	return nil
}

func (c Connections) Delete(name types.NamespacedName) {
	if cc, exist := c.index.Load(name); exist {
		var conn = cc.(Connection)
		if err := conn.Stop(); err != nil {
			log.Error(err, "failed to stop connection", "connection", conn.GetName())
		}
		c.index.Delete(name)
	}
}

func (c Connections) Put(conn Connection) {
	if cc, exist := c.index.LoadOrStore(conn.GetName(), conn); exist {
		var staleConn = cc.(Connection)
		if err := staleConn.Stop(); err != nil {
			log.Error(err, "failed to stop stable connection", "connection", staleConn.GetName())
		}
	}
}

func (c Connections) Cleanup() {
	c.index.Range(func(name, cc interface{}) bool {
		// delete
		c.index.Delete(name)
		// close connection
		var conn = cc.(Connection)
		if err := conn.Stop(); err != nil {
			log.Error(err, "failed to stop connection", "connection", conn.GetName())
		}
		return true
	})
}
