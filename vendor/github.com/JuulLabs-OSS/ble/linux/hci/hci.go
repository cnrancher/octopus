package hci

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/linux/hci/cmd"
	"github.com/JuulLabs-OSS/ble/linux/hci/evt"
	"github.com/JuulLabs-OSS/ble/linux/hci/socket"
	"github.com/pkg/errors"
)

// Command ...
type Command interface {
	OpCode() int
	Len() int
	Marshal([]byte) error
}

// CommandRP ...
type CommandRP interface {
	Unmarshal(b []byte) error
}

type handlerFn func(b []byte) error

type pkt struct {
	cmd  Command
	done chan []byte
}

// NewHCI returns a hci device.
func NewHCI(opts ...ble.Option) (*HCI, error) {
	h := &HCI{
		id: -1,

		chCmdPkt:  make(chan *pkt),
		chCmdBufs: make(chan []byte, 16),
		sent:      make(map[int]*pkt),
		muSent:    &sync.Mutex{},

		evth: map[int]handlerFn{},
		subh: map[int]handlerFn{},

		muConns:      &sync.Mutex{},
		conns:        make(map[uint16]*Conn),
		chMasterConn: make(chan *Conn),
		chSlaveConn:  make(chan *Conn),

		done: make(chan bool),
	}
	h.params.init()
	if err := h.Option(opts...); err != nil {
		return nil, errors.Wrap(err, "can't set options")
	}

	return h, nil
}

// HCI ...
type HCI struct {
	sync.Mutex

	params params

	skt io.ReadWriteCloser
	id  int

	// Host to Controller command flow control [Vol 2, Part E, 4.4]
	chCmdPkt  chan *pkt
	chCmdBufs chan []byte
	muSent    *sync.Mutex
	sent      map[int]*pkt

	// evtHub
	evth map[int]handlerFn
	subh map[int]handlerFn

	// aclHandler
	bufSize int
	bufCnt  int

	// Device information or status.
	addr    net.HardwareAddr
	txPwrLv int

	// adHist and adLast track the history of past scannable advertising packets.
	// Controller delivers AD(Advertising Data) and SR(Scan Response) separately
	// through HCI. Upon receiving an AD, no matter it's scannable or not, we
	// pass a Advertisement (AD only) to advHandler immediately.
	// Upon receiving a SR, we search the AD history for the AD from the same
	// device, and pass the Advertisiement (AD+SR) to advHandler.
	// The adHist and adLast are allocated in the Scan().
	advHandler ble.AdvHandler
	adHist     []*Advertisement
	adLast     int

	// Host to Controller Data Flow Control Packet-based Data flow control for LE-U [Vol 2, Part E, 4.1.1]
	// Minimum 27 bytes. 4 bytes of L2CAP Header, and 23 bytes Payload from upper layer (ATT)
	pool *Pool

	// L2CAP connections
	muConns      *sync.Mutex
	conns        map[uint16]*Conn
	chMasterConn chan *Conn // Dial returns master connections.
	chSlaveConn  chan *Conn // Peripheral accept slave connections.

	connectedHandler    func(evt.LEConnectionComplete)
	disconnectedHandler func(evt.DisconnectionComplete)

	dialerTmo   time.Duration
	listenerTmo time.Duration

	err  error
	done chan bool
}

// Init ...
func (h *HCI) Init() error {
	h.evth[0x3E] = h.handleLEMeta
	h.evth[evt.CommandCompleteCode] = h.handleCommandComplete
	h.evth[evt.CommandStatusCode] = h.handleCommandStatus
	h.evth[evt.DisconnectionCompleteCode] = h.handleDisconnectionComplete
	h.evth[evt.NumberOfCompletedPacketsCode] = h.handleNumberOfCompletedPackets

	h.subh[evt.LEAdvertisingReportSubCode] = h.handleLEAdvertisingReport
	h.subh[evt.LEConnectionCompleteSubCode] = h.handleLEConnectionComplete
	h.subh[evt.LEConnectionUpdateCompleteSubCode] = h.handleLEConnectionUpdateComplete
	h.subh[evt.LELongTermKeyRequestSubCode] = h.handleLELongTermKeyRequest
	// evt.EncryptionChangeCode:                     todo),
	// evt.ReadRemoteVersionInformationCompleteCode: todo),
	// evt.HardwareErrorCode:                        todo),
	// evt.DataBufferOverflowCode:                   todo),
	// evt.EncryptionKeyRefreshCompleteCode:         todo),
	// evt.AuthenticatedPayloadTimeoutExpiredCode:   todo),
	// evt.LEReadRemoteUsedFeaturesCompleteSubCode:   todo),
	// evt.LERemoteConnectionParameterRequestSubCode: todo),

	skt, err := socket.NewSocket(h.id)
	if err != nil {
		return err
	}
	h.skt = skt

	h.setAllowedCommands(1)

	go h.sktLoop()
	if err := h.init(); err != nil {
		return err
	}

	// Pre-allocate buffers with additional head room for lower layer headers.
	// HCI header (1 Byte) + ACL Data Header (4 bytes) + L2CAP PDU (or fragment)
	h.pool = NewPool(1+4+h.bufSize, h.bufCnt-1)

	h.Send(&h.params.advParams, nil)
	h.Send(&h.params.scanParams, nil)
	return nil
}

// Close ...
func (h *HCI) Close() error {
	return h.close(nil)
}

// Error ...
func (h *HCI) Error() error {
	return h.err
}

// Option sets the options specified.
func (h *HCI) Option(opts ...ble.Option) error {
	var err error
	for _, opt := range opts {
		err = opt(h)
	}
	return err
}

func (h *HCI) init() error {
	h.Send(&cmd.Reset{}, nil)

	ReadBDADDRRP := cmd.ReadBDADDRRP{}
	h.Send(&cmd.ReadBDADDR{}, &ReadBDADDRRP)

	a := ReadBDADDRRP.BDADDR
	h.addr = net.HardwareAddr([]byte{a[5], a[4], a[3], a[2], a[1], a[0]})

	ReadBufferSizeRP := cmd.ReadBufferSizeRP{}
	h.Send(&cmd.ReadBufferSize{}, &ReadBufferSizeRP)

	// Assume the buffers are shared between ACL-U and LE-U.
	h.bufCnt = int(ReadBufferSizeRP.HCTotalNumACLDataPackets)
	h.bufSize = int(ReadBufferSizeRP.HCACLDataPacketLength)

	LEReadBufferSizeRP := cmd.LEReadBufferSizeRP{}
	h.Send(&cmd.LEReadBufferSize{}, &LEReadBufferSizeRP)

	if LEReadBufferSizeRP.HCTotalNumLEDataPackets != 0 {
		// Okay, LE-U do have their own buffers.
		h.bufCnt = int(LEReadBufferSizeRP.HCTotalNumLEDataPackets)
		h.bufSize = int(LEReadBufferSizeRP.HCLEDataPacketLength)
	}

	LEReadAdvertisingChannelTxPowerRP := cmd.LEReadAdvertisingChannelTxPowerRP{}
	h.Send(&cmd.LEReadAdvertisingChannelTxPower{}, &LEReadAdvertisingChannelTxPowerRP)

	h.txPwrLv = int(LEReadAdvertisingChannelTxPowerRP.TransmitPowerLevel)

	LESetEventMaskRP := cmd.LESetEventMaskRP{}
	h.Send(&cmd.LESetEventMask{LEEventMask: 0x000000000000001F}, &LESetEventMaskRP)

	SetEventMaskRP := cmd.SetEventMaskRP{}
	h.Send(&cmd.SetEventMask{EventMask: 0x3dbff807fffbffff}, &SetEventMaskRP)

	WriteLEHostSupportRP := cmd.WriteLEHostSupportRP{}
	h.Send(&cmd.WriteLEHostSupport{LESupportedHost: 1, SimultaneousLEHost: 0}, &WriteLEHostSupportRP)

	LEWriteSuggDefaultDataLengthRP := cmd.LEWriteSuggDefaultDataLengthRP{}
	h.Send(&cmd.LEWriteSuggDefaultDataLength{MaxTxOctets: ble.MaxOctsDLE, MaxTxTime: ble.MaxTimeDLE},
		&LEWriteSuggDefaultDataLengthRP)

	return h.err
}

// Send ...
func (h *HCI) Send(c Command, r CommandRP) error {
	b, err := h.send(c)
	if err != nil {
		return err
	}
	if len(b) > 0 && b[0] != 0x00 {
		return ErrCommand(b[0])
	}
	if r != nil {
		return r.Unmarshal(b)
	}
	return nil
}

func (h *HCI) send(c Command) ([]byte, error) {
	if h.err != nil {
		return nil, h.err
	}
	p := &pkt{c, make(chan []byte)}
	b := <-h.chCmdBufs
	b[0] = byte(pktTypeCommand) // HCI header
	b[1] = byte(c.OpCode())
	b[2] = byte(c.OpCode() >> 8)
	b[3] = byte(c.Len())
	if err := c.Marshal(b[4:]); err != nil {
		h.close(fmt.Errorf("hci: failed to marshal cmd"))
	}

	h.muSent.Lock()
	h.sent[c.OpCode()] = p
	h.muSent.Unlock()
	if n, err := h.skt.Write(b[:4+c.Len()]); err != nil {
		h.close(fmt.Errorf("hci: failed to send cmd"))
	} else if n != 4+c.Len() {
		h.close(fmt.Errorf("hci: failed to send whole cmd pkt to hci socket"))
	}

	var ret []byte
	var err error

	// emergency timeout to prevent calls from locking up if the HCI
	// interface doesn't respond.  Responsed here should normally be fast
	// a timeout indicates a major problem with HCI.
	timeout := time.NewTimer(10 * time.Second)
	select {
	case <-timeout.C:
		err = fmt.Errorf("hci: no response to command, hci connection failed")
		ret = nil
	case <-h.done:
		err = h.err
		ret = nil
	case b := <-p.done:
		err = nil
		ret = b
	}
	timeout.Stop()

	// clear sent table when done, we sometimes get command complete or
	// command status messages with no matching send, which can attempt to
	// access stale packets in sent and fail or lock up.
	h.muSent.Lock()
	delete(h.sent, c.OpCode())
	h.muSent.Unlock()

	return ret, err
}

func (h *HCI) sktLoop() {
	b := make([]byte, 4096)
	defer close(h.done)
	for {
		n, err := h.skt.Read(b)
		if n == 0 || err != nil {
			if err == io.EOF {
				h.err = err //callers depend on detecting io.EOF, don't wrap it.
			} else {
				h.err = fmt.Errorf("skt: %s", err)
			}
			return
		}
		p := make([]byte, n)
		copy(p, b)
		if err := h.handlePkt(p); err != nil {
			// Some bluetooth devices may append vendor specific packets at the last,
			// in this case, simply ignore them.
			if strings.HasPrefix(err.Error(), "unsupported vendor packet:") {
				_ = logger.Error("skt: %v", err)
			} else {
				log.Printf("skt: %v", err)
				continue
			}
		}
	}
}

func (h *HCI) close(err error) error {
	h.err = err
	if h.skt != nil {
		return h.skt.Close()
	}
	return err
}

func (h *HCI) handlePkt(b []byte) error {
	// Strip the 1-byte HCI header and pass down the rest of the packet.
	t, b := b[0], b[1:]
	switch t {
	case pktTypeCommand:
		return fmt.Errorf("unmanaged cmd: % X", b)
	case pktTypeACLData:
		return h.handleACL(b)
	case pktTypeSCOData:
		return fmt.Errorf("unsupported sco packet: % X", b)
	case pktTypeEvent:
		return h.handleEvt(b)
	case pktTypeVendor:
		return fmt.Errorf("unsupported vendor packet: % X", b)
	default:
		return fmt.Errorf("invalid packet: 0x%02X % X", t, b)
	}
}

func (h *HCI) handleACL(b []byte) error {
	handle := packet(b).handle()
	h.muConns.Lock()
	c, ok := h.conns[handle]
	h.muConns.Unlock()
	if !ok {
		_ = logger.Warn("invalid connection handle on ACL packet", "handle", handle)
		return nil
	}
	c.chInPkt <- b
	return nil
}

func (h *HCI) handleEvt(b []byte) error {
	code, plen := int(b[0]), int(b[1])
	if plen != len(b[2:]) {
		return fmt.Errorf("invalid event packet: % X", b)
	}
	if code == evt.CommandCompleteCode || code == evt.CommandStatusCode {
		if f := h.evth[code]; f != nil {
			return f(b[2:])
		}
	}
	if plen != len(b[2:]) {
		h.err = fmt.Errorf("invalid event packet: % X", b)
	}
	if f := h.evth[code]; f != nil {
		h.err = f(b[2:])
		return nil
	}
	if code == 0xff { // Ignore vendor events
		return nil
	}
	return fmt.Errorf("unsupported event packet: % X", b)
}

func (h *HCI) handleLEMeta(b []byte) error {
	subcode := int(b[0])
	if f := h.subh[subcode]; f != nil {
		return f(b)
	}
	return fmt.Errorf("unsupported LE event: % X", b)
}

func (h *HCI) handleLEAdvertisingReport(b []byte) error {
	if h.advHandler == nil {
		return nil
	}

	e := evt.LEAdvertisingReport(b)
	for i := 0; i < int(e.NumReports()); i++ {
		var a *Advertisement
		switch e.EventType(i) {
		case evtTypAdvInd:
			fallthrough
		case evtTypAdvScanInd:
			a = newAdvertisement(e, i)
			h.adHist[h.adLast] = a
			h.adLast++
			if h.adLast == len(h.adHist) {
				h.adLast = 0
			}
		case evtTypScanRsp:
			sr := newAdvertisement(e, i)
			for idx := h.adLast - 1; idx != h.adLast; idx-- {
				if idx == -1 {
					idx = len(h.adHist) - 1
				}
				if h.adHist[idx] == nil {
					break
				}
				if h.adHist[idx].Addr().String() == sr.Addr().String() {
					h.adHist[idx].setScanResponse(sr)
					a = h.adHist[idx]
					break
				}
			}
			// Got a SR without having received an associated AD before?
			if a == nil {
				return fmt.Errorf("received scan response %s with no associated Advertising Data packet", sr.Addr())
			}
		default:
			a = newAdvertisement(e, i)
		}
		go h.advHandler(a)
	}

	return nil
}

func (h *HCI) handleCommandComplete(b []byte) error {
	e := evt.CommandComplete(b)
	h.setAllowedCommands(int(e.NumHCICommandPackets()))

	// NOP command, used for flow control purpose [Vol 2, Part E, 4.4]
	// no handling other than setAllowedCommands needed
	if e.CommandOpcode() == 0x0000 {
		return nil
	}
	h.muSent.Lock()
	p, found := h.sent[int(e.CommandOpcode())]
	h.muSent.Unlock()
	if !found {
		return fmt.Errorf("can't find the cmd for CommandCompleteEP: % X", e)
	}
	p.done <- e.ReturnParameters()
	return nil
}

func (h *HCI) handleCommandStatus(b []byte) error {
	e := evt.CommandStatus(b)
	h.setAllowedCommands(int(e.NumHCICommandPackets()))

	h.muSent.Lock()
	p, found := h.sent[int(e.CommandOpcode())]
	h.muSent.Unlock()
	if !found {
		return fmt.Errorf("can't find the cmd for CommandStatusEP: % X", e)
	}
	p.done <- []byte{e.Status()}
	return nil
}

func (h *HCI) handleLEConnectionComplete(b []byte) error {
	e := evt.LEConnectionComplete(b)
	c := newConn(h, e)
	h.muConns.Lock()
	h.conns[e.ConnectionHandle()] = c
	h.muConns.Unlock()
	if e.Role() == roleMaster {
		if e.Status() == 0x00 {
			select {
			case h.chMasterConn <- c:
			default:
				go c.Close()
			}
			return nil
		}
		if ErrCommand(e.Status()) == ErrConnID {
			// The connection was canceled successfully.
			return nil
		}
		return nil
	}
	if e.Status() == 0x00 {
		h.chSlaveConn <- c
		// When a controller accepts a connection, it moves from advertising
		// state to idle/ready state. Host needs to explicitly ask the
		// controller to re-enable advertising. Note that the host was most
		// likely in advertising state. Otherwise it couldn't accept the
		// connection in the first place. The only exception is that user
		// asked the host to stop advertising during this tiny window.
		// The re-enabling might failed or ignored by the controller, if
		// it had reached the maximum number of concurrent connections.
		// So we also re-enable the advertising when a connection disconnected
		h.params.RLock()
		if h.params.advEnable.AdvertisingEnable == 1 {
			go h.Send(&cmd.LESetAdvertiseEnable{0}, nil)
		}
		h.params.RUnlock()
	}
	if h.connectedHandler != nil {
		h.connectedHandler(e)
	}
	return nil
}

func (h *HCI) handleLEConnectionUpdateComplete(b []byte) error {
	return nil
}

func (h *HCI) handleDisconnectionComplete(b []byte) error {
	e := evt.DisconnectionComplete(b)
	h.muConns.Lock()
	c, found := h.conns[e.ConnectionHandle()]
	delete(h.conns, e.ConnectionHandle())
	h.muConns.Unlock()
	if !found {
		return fmt.Errorf("disconnecting an invalid handle %04X", e.ConnectionHandle())
	}
	close(c.chInPkt)

	if c.param.Role() == roleSlave {
		// Re-enable advertising, if it was advertising. Refer to the
		// handleLEConnectionComplete() for details.
		// This may failed with ErrCommandDisallowed, if the controller
		// was actually in advertising state. It does no harm though.
		h.params.RLock()
		if h.params.advEnable.AdvertisingEnable == 1 {
			go h.Send(&h.params.advEnable, nil)
		}
		h.params.RUnlock()
	} else {
		// remote peripheral disconnected
		close(c.chDone)
	}
	// When a connection disconnects, all the sent packets and weren't acked yet
	// will be recycled. [Vol2, Part E 4.1.1]
	//
	// must be done with the pool locked to avoid race conditions where
	// writePDU is in progress and does a Get from the pool after this completes,
	// leaking a buffer from the main pool.
	c.txBuffer.LockPool()
	c.txBuffer.PutAll()
	c.txBuffer.UnlockPool()
	if h.disconnectedHandler != nil {
		h.disconnectedHandler(e)
	}
	return nil
}

func (h *HCI) handleNumberOfCompletedPackets(b []byte) error {
	e := evt.NumberOfCompletedPackets(b)
	h.muConns.Lock()
	defer h.muConns.Unlock()
	for i := 0; i < int(e.NumberOfHandles()); i++ {
		c, found := h.conns[e.ConnectionHandle(i)]
		if !found {
			continue
		}

		// Put the delivered buffers back to the pool.
		for j := 0; j < int(e.HCNumOfCompletedPackets(i)); j++ {
			c.txBuffer.Put()
		}
	}
	return nil
}

func (h *HCI) handleLELongTermKeyRequest(b []byte) error {
	e := evt.LELongTermKeyRequest(b)
	return h.Send(&cmd.LELongTermKeyRequestNegativeReply{
		ConnectionHandle: e.ConnectionHandle(),
	}, nil)
}

func (h *HCI) setAllowedCommands(n int) {

	//hard-coded limit to command queue depth
	//matches make(chan []byte, 16) in NewHCI
	// TODO make this a constant, decide correct size
	if n > 16 {
		n = 16
	}

	for len(h.chCmdBufs) < n {
		h.chCmdBufs <- make([]byte, 64) // TODO make buffer size a constant
	}
}
