// pmgrdlg.go: Implements the PeripheralManagerDelegate interface.
// CoreBluetooth communicates events asynchronously via callbacks.  This file
// implements a synchronous interface by translating these callbacks into
// channel operations.

package darwin

import (
	"bytes"
	"fmt"
	"log"

	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/cbgo"
)

func (d *Device) PeripheralManagerDidUpdateState(pmgr cbgo.PeripheralManager) {
	d.evl.stateChanged.RxSignal(struct{}{})
}

func (d *Device) DidAddService(pmgr cbgo.PeripheralManager, svc cbgo.Service, err error) {
	d.evl.svcAdded.RxSignal(err)
}

func (d *Device) DidStartAdvertising(pmgr cbgo.PeripheralManager, err error) {
	d.evl.advStarted.RxSignal(err)
}

func (d *Device) DidReceiveReadRequest(pmgr cbgo.PeripheralManager, cbreq cbgo.ATTRequest) {
	chr, _ := d.pc.findChr(cbreq.Characteristic())
	if chr == nil || chr.ReadHandler == nil {
		return
	}

	c := d.findConn(cbreq.Central().Identifier())
	if c == nil {
		var err error
		c, err = newPeripheralConn(d, cbreq.Central())
		if err != nil {
			log.Printf("failed to process read response: %v", err)
			return
		}
	}

	req := ble.NewRequest(c, nil, cbreq.Offset())
	buf := bytes.NewBuffer(make([]byte, 0, c.txMTU-1))
	rsp := ble.NewResponseWriter(buf)
	chr.ReadHandler.ServeRead(req, rsp)
	cbreq.SetValue(buf.Bytes())

	pmgr.RespondToRequest(cbreq, cbgo.ATTError(rsp.Status()))
}

func (d *Device) DidReceiveWriteRequests(pmgr cbgo.PeripheralManager, cbreqs []cbgo.ATTRequest) {
	serveOne := func(cbreq cbgo.ATTRequest) {
		chr, _ := d.pc.findChr(cbreq.Characteristic())
		if chr == nil || chr.WriteHandler == nil {
			return
		}

		c := d.findConn(cbreq.Central().Identifier())
		if c == nil {
			var err error
			c, err = newPeripheralConn(d, cbreq.Central())
			if err != nil {
				log.Printf("failed to process write response: %v", err)
				return
			}
		}

		req := ble.NewRequest(c, cbreq.Value(), cbreq.Offset())
		rsp := ble.NewResponseWriter(nil)
		chr.WriteHandler.ServeWrite(req, rsp)

		pmgr.RespondToRequest(cbreq, cbgo.ATTError(rsp.Status()))
	}

	for _, cbreq := range cbreqs {
		serveOne(cbreq)
	}
}

func (d *Device) CentralDidSubscribe(pmgr cbgo.PeripheralManager, cent cbgo.Central, cbchr cbgo.Characteristic) {
	c := d.findConn(cent.Identifier())
	if c == nil {
		var err error
		c, err = newPeripheralConn(d, cent)
		if err != nil {
			log.Printf("failed to process subscribe request: %v", err)
			return
		}
	}

	if c.notifiers[cbchr] != nil {
		return
	}

	chr, _ := d.pc.findChr(cbchr)
	if chr == nil {
		return
	}

	send := func(b []byte) (int, error) {
		sent := d.pm.UpdateValue(b, cbchr, nil)
		if !sent {
			return len(b), fmt.Errorf("failed to send notification: tx queue full")
		}

		return len(b), nil
	}
	n := ble.NewNotifier(send)
	c.notifiers[cbchr] = n
	req := ble.NewRequest(c, nil, 0) // convey *conn to user handler.

	go chr.NotifyHandler.ServeNotify(req, n)
}

func (d *Device) CentralDidUnsubscribe(pmgr cbgo.PeripheralManager, cent cbgo.Central, chr cbgo.Characteristic) {
	c := d.findConn(cent.Identifier())
	if c == nil {
		var err error
		c, err = newPeripheralConn(d, cent)
		if err != nil {
			log.Printf("failed to process unsubscribe request: %v", err)
			return
		}
	}

	n := c.notifiers[chr]
	if n != nil {
		if err := n.Close(); err != nil {
			log.Printf("failed to close notifier: %v", err)
		}
		delete(c.notifiers, chr)
	}
}
