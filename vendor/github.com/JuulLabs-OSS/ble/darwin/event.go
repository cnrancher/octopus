package darwin

import (
	"sync"
)

// eventSlot is a receiver for asynchronous events from CoreBluetooth.  To
// prevent deadlock in the case of spurious events, eventSlot discards incoming
// signals if it is not explicitly listening for them.
type eventSlot struct {
	ch  chan interface{}
	mtx sync.Mutex
}

func (e *eventSlot) closeNoLock() {
	if e.ch == nil {
		return
	}

	// Drain channel.
	for len(e.ch) > 0 {
		<-e.ch
	}

	close(e.ch)
	e.ch = nil
}

func (e *eventSlot) Close() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	e.closeNoLock()
}

// Listen listens for a single event on this slot.  It returns the channel on
// which the event will be received.
func (e *eventSlot) Listen() chan interface{} {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.ch != nil {
		e.closeNoLock()
	}

	e.ch = make(chan interface{})
	return e.ch
}

// RxSignal causes the event slot to process the given signal (i.e., it sends a
// signal to the slot).  It blocks until the signal is consumed by a client or
// until the slot is closed.
func (e *eventSlot) RxSignal(sig interface{}) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.ch == nil {
		// Not listening.  Discard signal.
		return
	}

	e.ch <- sig

	// Stop listening.
	e.closeNoLock()
}

type eventConnected struct {
	conn *conn
	err  error
}

type eventRSSIRead struct {
	rssi int
	err  error
}

// Each Client owns one of these (us-as-central).
type clientEventListener struct {
	svcsDiscovered eventSlot // error
	chrsDiscovered eventSlot // error
	dscsDiscovered eventSlot // error
	chrWritten     eventSlot // error
	dscRead        eventSlot // error
	dscWritten     eventSlot // error
	notifyChanged  eventSlot // error
	rssiRead       eventSlot // *eventRSSIRead
}

func (cevl *clientEventListener) Close() {
	cevl.svcsDiscovered.Close()
	cevl.chrsDiscovered.Close()
	cevl.dscsDiscovered.Close()
	cevl.chrWritten.Close()
	cevl.dscRead.Close()
	cevl.dscWritten.Close()
	cevl.notifyChanged.Close()
	cevl.rssiRead.Close()
}

// Each Device owns one of these (us-as-peripheral).
type deviceEventListener struct {
	stateChanged eventSlot // struct{}
	connected    eventSlot // *eventConnected
	svcAdded     eventSlot // error
	advStarted   eventSlot // error
}

func (devl *deviceEventListener) Close() {
	devl.stateChanged.Close()
	devl.svcAdded.Close()
	devl.advStarted.Close()
}
