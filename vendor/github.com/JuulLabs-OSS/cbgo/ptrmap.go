package cbgo

import (
	"sync"
	"unsafe"
)

// ptrMap is a thread-safe map with pointer keys and generic values.
type ptrMap struct {
	m   map[unsafe.Pointer]interface{}
	mtx sync.RWMutex
}

func newPtrMap() *ptrMap {
	return &ptrMap{
		m: map[unsafe.Pointer]interface{}{},
	}
}

func (pm *ptrMap) find(p unsafe.Pointer) interface{} {
	pm.mtx.RLock()
	defer pm.mtx.RUnlock()

	return pm.m[p]
}

func (pm *ptrMap) add(p unsafe.Pointer, itf interface{}) error {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	pm.m[p] = itf

	return nil
}

func (pm *ptrMap) del(p unsafe.Pointer) {
	pm.mtx.Lock()
	defer pm.mtx.Unlock()

	delete(pm.m, p)
}
