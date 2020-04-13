package critical

import (
	"context"
)

func Context(stop <-chan struct{}, beforeStopFns ...func()) context.Context {
	var ctx, cancel = context.WithCancel(context.Background())
	var closeFn = func() {
		if len(beforeStopFns) != 0 {
			for _, fn := range beforeStopFns {
				fn()
			}
		}
		cancel()
	}
	go func() {
		for {
			select {
			case <-stop:
				closeFn()
				return
			}
		}
	}()
	return ctx
}
