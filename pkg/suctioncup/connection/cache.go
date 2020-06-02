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

// Get returns a connection by name
func (c Connections) Get(name types.NamespacedName) Connection {
	if cc, exist := c.index.Load(name); exist {
		return cc.(Connection)
	}
	return nil
}

// Delete deletes the connection by name, the return value represents whether there is a deleted target.
func (c Connections) Delete(name types.NamespacedName) (exist bool) {
	if cc, exist := c.index.Load(name); exist {
		var conn = cc.(Connection)
		if err := conn.Stop(); err != nil {
			log.Error(err, "Failed to stop connection", "connection", conn.GetName())
		}
		c.index.Delete(name)
		return true
	}
	return false
}

// Put puts a connection in index, then return value represents whether to overwrite an existing connection.
func (c Connections) Put(conn Connection) (overwrite bool) {
	if cc, exist := c.index.LoadOrStore(conn.GetName(), conn); exist {
		var staleConn = cc.(Connection)
		if err := staleConn.Stop(); err != nil {
			log.Error(err, "Failed to stop stable connection", "connection", staleConn.GetName())
		}
		c.index.Store(conn.GetName(), conn)
		return true
	}
	return false
}

// Cleanup cleans all connections of index
func (c Connections) Cleanup() {
	c.index.Range(func(name, cc interface{}) bool {
		// delete
		c.index.Delete(name)
		// close connection
		var conn = cc.(Connection)
		if err := conn.Stop(); err != nil {
			log.Error(err, "Failed to stop connection", "connection", conn.GetName())
		}
		return true
	})
}
