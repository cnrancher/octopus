package darwin

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/cbgo"

	"sync"
)

type connectResult struct {
	conn *conn
	err  error
}

// Device is either a Peripheral or Central device.
type Device struct {
	// Embed these two bases so we don't have to override all the esoteric
	// functions defined by CoreBluetooth delegate interfaces.
	cbgo.CentralManagerDelegateBase
	cbgo.PeripheralManagerDelegateBase

	cm  cbgo.CentralManager
	pm  cbgo.PeripheralManager
	evl deviceEventListener
	pc  profCache

	conns    map[string]*conn
	connLock sync.Mutex

	advHandler ble.AdvHandler
}

// NewDevice returns a BLE device.
func NewDevice(opts ...ble.Option) (*Device, error) {
	d := &Device{
		cm:    cbgo.NewCentralManager(nil),
		pm:    cbgo.NewPeripheralManager(nil),
		pc:    newProfCache(),
		conns: make(map[string]*conn),
	}

	// Only proceed if Bluetooth is enabled.

	blockUntilStateChange := func(getState func() cbgo.ManagerState) {
		if getState() != cbgo.ManagerStateUnknown {
			return
		}

		// Wait until state changes or until one second passes (whichever
		// happens first).
		for {
			select {
			case <-d.evl.stateChanged.Listen():
				if getState() != cbgo.ManagerStateUnknown {
					return
				}

			case <-time.NewTimer(time.Second).C:
				return
			}
		}
	}

	// Ensure central manager is ready.
	d.cm.SetDelegate(d)
	blockUntilStateChange(d.cm.State)
	if d.cm.State() != cbgo.ManagerStatePoweredOn {
		return nil, fmt.Errorf("central manager has invalid state: have=%d want=%d: is Bluetooth turned on?",
			d.cm.State(), cbgo.ManagerStatePoweredOn)
	}

	// Ensure peripheral manager is ready.
	d.pm.SetDelegate(d)
	blockUntilStateChange(d.pm.State)
	if d.pm.State() != cbgo.ManagerStatePoweredOn {
		return nil, fmt.Errorf("peripheral manager has invalid state: have=%d want=%d: is Bluetooth turned on?",
			d.pm.State(), cbgo.ManagerStatePoweredOn)
	}

	return d, nil
}

// Option sets the options specified.
func (d *Device) Option(opts ...ble.Option) error {
	return nil
}

// Scan ...
func (d *Device) Scan(ctx context.Context, allowDup bool, h ble.AdvHandler) error {
	d.advHandler = h

	d.cm.Scan(nil, &cbgo.CentralManagerScanOpts{
		AllowDuplicates: allowDup,
	})

	<-ctx.Done()
	d.cm.StopScan()

	return ctx.Err()
}

// Dial ...
func (d *Device) Dial(ctx context.Context, a ble.Addr) (ble.Client, error) {
	uuid, err := cbgo.ParseUUID(uuidStrWithDashes(a.String()))
	if err != nil {
		return nil, fmt.Errorf("dial failed: invalid peer address: %s", a)
	}

	prphs := d.cm.RetrievePeripheralsWithIdentifiers([]cbgo.UUID{uuid})
	if len(prphs) == 0 {
		return nil, fmt.Errorf("dial failed: no peer with address: %s", a)
	}

	ch := d.evl.connected.Listen()
	defer d.evl.connected.Close()

	d.cm.Connect(prphs[0], nil)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case itf := <-ch:
		if itf == nil {
			return nil, fmt.Errorf("connect failed: aborted")
		}

		ev := itf.(*eventConnected)
		if ev.err != nil {
			return nil, ev.err
		} else {
			ev.conn.SetContext(ctx)
			return NewClient(d.cm, ev.conn)
		}
	}
}

// Stop ...
func (d *Device) Stop() error {
	return nil
}

func (d *Device) closeConns() {
	d.connLock.Lock()
	defer d.connLock.Unlock()

	for _, c := range d.conns {
		c.Close()
	}
}

func (d *Device) findConn(a ble.Addr) *conn {
	d.connLock.Lock()
	defer d.connLock.Unlock()

	return d.conns[a.String()]
}

func (d *Device) addConn(c *conn) error {
	d.connLock.Lock()
	defer d.connLock.Unlock()

	if d.conns[c.addr.String()] != nil {
		return fmt.Errorf("failed to add connection: already exists: addr=%v", c.addr)
	}

	d.conns[c.addr.String()] = c

	return nil
}

func (d *Device) delConn(a ble.Addr) {
	d.connLock.Lock()
	defer d.connLock.Unlock()

	delete(d.conns, a.String())
}

func (d *Device) connectFail(err error) {
	d.evl.connected.RxSignal(&eventConnected{
		err: err,
	})
}

func chrPropPerm(c *ble.Characteristic) (cbgo.CharacteristicProperties, cbgo.AttributePermissions) {
	var prop cbgo.CharacteristicProperties
	var perm cbgo.AttributePermissions

	if c.Property&ble.CharRead != 0 {
		prop |= cbgo.CharacteristicPropertyRead
		if ble.CharRead&c.Secure != 0 {
			perm |= cbgo.AttributePermissionsReadEncryptionRequired
		} else {
			perm |= cbgo.AttributePermissionsReadable
		}
	}
	if c.Property&ble.CharWriteNR != 0 {
		prop |= cbgo.CharacteristicPropertyWriteWithoutResponse
		if c.Secure&ble.CharWriteNR != 0 {
			perm |= cbgo.AttributePermissionsWriteEncryptionRequired
		} else {
			perm |= cbgo.AttributePermissionsWriteable
		}
	}
	if c.Property&ble.CharWrite != 0 {
		prop |= cbgo.CharacteristicPropertyWrite
		if c.Secure&ble.CharWrite != 0 {
			perm |= cbgo.AttributePermissionsWriteEncryptionRequired
		} else {
			perm |= cbgo.AttributePermissionsWriteable
		}
	}
	if c.Property&ble.CharNotify != 0 {
		if c.Secure&ble.CharNotify != 0 {
			prop |= cbgo.CharacteristicPropertyNotifyEncryptionRequired
		} else {
			prop |= cbgo.CharacteristicPropertyNotify
		}
	}
	if c.Property&ble.CharIndicate != 0 {
		if c.Secure&ble.CharIndicate != 0 {
			prop |= cbgo.CharacteristicPropertyIndicateEncryptionRequired
		} else {
			prop |= cbgo.CharacteristicPropertyIndicate
		}
	}

	return prop, perm
}

func (d *Device) AddService(svc *ble.Service) error {
	chrMap := make(map[*ble.Characteristic]cbgo.Characteristic)
	dscMap := make(map[*ble.Descriptor]cbgo.Descriptor)

	msvc := cbgo.NewMutableService(cbgo.UUID(svc.UUID), true)

	var mchrs []cbgo.MutableCharacteristic
	for _, c := range svc.Characteristics {
		prop, perm := chrPropPerm(c)
		mchr := cbgo.NewMutableCharacteristic(cbgo.UUID(c.UUID), prop, c.Value, perm)

		var mdscs []cbgo.MutableDescriptor
		for _, d := range c.Descriptors {
			mdsc := cbgo.NewMutableDescriptor(cbgo.UUID(d.UUID), d.Value)
			mdscs = append(mdscs, mdsc)
			dscMap[d] = mdsc.Descriptor()
		}
		mchr.SetDescriptors(mdscs)

		mchrs = append(mchrs, mchr)
		chrMap[c] = mchr.Characteristic()
	}
	msvc.SetCharacteristics(mchrs)

	ch := d.evl.svcAdded.Listen()
	d.pm.AddService(msvc)

	itf := <-ch
	if itf != nil {
		return itf.(error)
	}

	d.pc.addSvc(svc, msvc.Service())
	for chr, cbc := range chrMap {
		d.pc.addChr(chr, cbc)
	}
	for dsc, cbd := range dscMap {
		d.pc.addDsc(dsc, cbd)
	}

	return nil
}

func (d *Device) RemoveAllServices() error {
	d.pm.RemoveAllServices()
	return nil
}

func (d *Device) SetServices(svcs []*ble.Service) error {
	d.RemoveAllServices()
	for _, s := range svcs {
		d.AddService(s)
	}

	return nil
}

func (d *Device) stopAdvertising() error {
	d.pm.StopAdvertising()
	return nil
}

func (d *Device) advData(ctx context.Context, ad cbgo.AdvData) error {
	ch := d.evl.advStarted.Listen()
	d.pm.StartAdvertising(ad)

	itf := <-ch
	if itf != nil {
		return itf.(error)
	}

	<-ctx.Done()
	_ = d.stopAdvertising()
	return ctx.Err()
}

func (d *Device) Advertise(ctx context.Context, adv ble.Advertisement) error {
	ad := cbgo.AdvData{}

	ad.LocalName = adv.LocalName()
	for _, u := range adv.Services() {
		ad.ServiceUUIDs = append(ad.ServiceUUIDs, cbgo.UUID(u))
	}

	return d.advData(ctx, ad)
}

func (d *Device) AdvertiseNameAndServices(ctx context.Context, name string, uuids ...ble.UUID) error {
	a := &adv{
		localName: name,
		svcUUIDs:  uuids,
	}

	return d.Advertise(ctx, a)
}

func (d *Device) AdvertiseMfgData(ctx context.Context, id uint16, b []byte) error {
	// CoreBluetooth doesn't let you specify manufacturer data :(
	return errors.New("Not supported")
}

func (d *Device) AdvertiseServiceData16(ctx context.Context, id uint16, b []byte) error {
	// CoreBluetooth doesn't let you specify service data :(
	return errors.New("Not supported")
}

func (d *Device) AdvertiseIBeaconData(ctx context.Context, b []byte) error {
	ad := cbgo.AdvData{
		IBeaconData: b,
	}
	return d.advData(ctx, ad)
}

func (d *Device) AdvertiseIBeacon(ctx context.Context, u ble.UUID, major, minor uint16, pwr int8) error {
	b := make([]byte, 21)
	copy(b, ble.Reverse(u))                   // Big endian
	binary.BigEndian.PutUint16(b[16:], major) // Big endian
	binary.BigEndian.PutUint16(b[18:], minor) // Big endian
	b[20] = uint8(pwr)                        // Measured Tx Power
	return d.AdvertiseIBeaconData(ctx, b)
}
