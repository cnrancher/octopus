package critical

func Chan(stop <-chan struct{}, beforeStopFns ...func()) <-chan struct{} {
	var next = make(chan struct{})
	var closeFn = func() {
		if len(beforeStopFns) != 0 {
			for _, fn := range beforeStopFns {
				fn()
			}
		}
		close(next)
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
	return next
}
