package safechan

import (
	"sync"
)

// SafeCloseChannel wraps a channel to avoid duplicate close
type SafeCloseChannel struct {
	C    chan struct{}
	once sync.Once
}

// NewSafeCloseChannel create a SafeCloseChannel instance
func NewSafeCloseChannel() *SafeCloseChannel {
	return &SafeCloseChannel{C: make(chan struct{})}
}

// Close closes the wrapped channel only once
func (scc *SafeCloseChannel) Close() {
	scc.once.Do(func() {
		close(scc.C)
	})
}
