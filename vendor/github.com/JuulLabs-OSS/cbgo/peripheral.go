package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import (
	"unsafe"
)

// PeripheralState: https://developer.apple.com/documentation/corebluetooth/cbperipheralstate
type PeripheralState int

const (
	PeripheralStateDisconnected  = PeripheralState(C.CBPeripheralStateDisconnected)
	PeripheralStateConnecting    = PeripheralState(C.CBPeripheralStateConnecting)
	PeripheralStateConnected     = PeripheralState(C.CBPeripheralStateConnected)
	PeripheralStateDisconnecting = PeripheralState(C.CBPeripheralStateDisconnecting)
)

var prphPtrMap = newPtrMap()

// Peripheral: https://developer.apple.com/documentation/corebluetooth/cbperipheral
type Peripheral struct {
	ptr unsafe.Pointer
}

func findPeripheralDlg(ptr unsafe.Pointer) PeripheralDelegate {
	itf := prphPtrMap.find(ptr)
	if itf == nil {
		return nil
	}

	return itf.(PeripheralDelegate)
}

func addPeripheralDlg(ptr unsafe.Pointer, dlg PeripheralDelegate) {
	prphPtrMap.add(ptr, dlg)
}

// Identifier: https://developer.apple.com/documentation/corebluetooth/cbpeer/1620687-identifier
func (p Peripheral) Identifier() UUID {
	cstr := C.cb_peer_identifier(p.ptr)
	return MustParseUUID(C.GoString(cstr))
}

// Name: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519029-name
func (p Peripheral) Name() string {
	cstr := C.cb_prph_name(p.ptr)
	return C.GoString(cstr)
}

// SetDelegate: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518730-delegate
func (p Peripheral) SetDelegate(dlg PeripheralDelegate) {
	addPeripheralDlg(p.ptr, dlg)
	C.cb_prph_set_delegate(p.ptr, C.bool(dlg != nil))
}

// DiscoverServices: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518706-discoverservices
func (p Peripheral) DiscoverServices(svcUUIDs []UUID) {
	arr := uuidsToStrArr(svcUUIDs)
	defer freeStrArr(&arr)

	C.cb_prph_discover_svcs(p.ptr, &arr)
}

// DiscoverIncludedServices: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519014-discoverincludedservices
func (p Peripheral) DiscoverIncludedServices(svcUUIDs []UUID, svc Service) {
	arr := uuidsToStrArr(svcUUIDs)
	defer freeStrArr(&arr)

	C.cb_prph_discover_included_svcs(p.ptr, &arr, svc.ptr)
}

// Services: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518978-services
func (p Peripheral) Services() []Service {
	oa := C.cb_prph_services(p.ptr)
	defer C.free(unsafe.Pointer(oa.objs))

	svcs := make([]Service, oa.count)
	for i, _ := range svcs {
		obj := getObjArrElem(&oa, i)
		svcs[i] = Service{ptr: obj}
	}

	return svcs
}

// DiscoverCharacteristics: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518797-discovercharacteristics
func (p Peripheral) DiscoverCharacteristics(chrUUIDs []UUID, svc Service) {
	arr := uuidsToStrArr(chrUUIDs)
	defer freeStrArr(&arr)

	C.cb_prph_discover_chrs(p.ptr, svc.ptr, &arr)
}

// DiscoverDescriptors: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519070-discoverdescriptorsforcharacteri
func (p Peripheral) DiscoverDescriptors(chr Characteristic) {
	C.cb_prph_discover_dscs(p.ptr, chr.ptr)
}

// ReadCharacteristic: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518759-readvalueforcharacteristic
func (p Peripheral) ReadCharacteristic(chr Characteristic) {
	C.cb_prph_read_chr(p.ptr, chr.ptr)
}

// ReadDescriptor: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518789-readvaluefordescriptor
func (p Peripheral) ReadDescriptor(dsc Descriptor) {
	C.cb_prph_read_dsc(p.ptr, dsc.ptr)
}

// WriteCharacteristic: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518747-writevalue
func (p Peripheral) WriteCharacteristic(data []byte, chr Characteristic, withRsp bool) {
	ba := byteSliceToByteArr(data)
	defer C.free(unsafe.Pointer(ba.data))

	C.cb_prph_write_chr(p.ptr, chr.ptr, &ba, chrWriteType(withRsp))
}

// WriteDescriptor: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519107-writevalue
func (p Peripheral) WriteDescriptor(data []byte, dsc Descriptor) {
	ba := byteSliceToByteArr(data)
	defer C.free(unsafe.Pointer(ba.data))

	C.cb_prph_write_dsc(p.ptr, dsc.ptr, &ba)
}

// MaximumWriteValueLength: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1620312-maximumwritevaluelengthfortype
func (p Peripheral) MaximumWriteValueLength(withRsp bool) int {
	val := C.cb_prph_max_write_len(p.ptr, chrWriteType(withRsp))
	return int(val)
}

// SetNotify: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1518949-setnotifyvalue
func (p Peripheral) SetNotify(enabled bool, chr Characteristic) {
	C.cb_prph_set_notify(p.ptr, C.bool(enabled), chr.ptr)
}

// PeripheralState: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519113-state
func (p Peripheral) State() PeripheralState {
	return PeripheralState(C.cb_prph_state(p.ptr))
}

// CanSendWriteWithoutResponse: https://developer.apple.com/documentation/corebluetooth/cbperipheral/2891512-cansendwritewithoutresponse
func (p Peripheral) CanSendWriteWithoutResponse() bool {
	return bool(C.cb_prph_can_send_write_without_rsp(p.ptr))
}

// ReadRSSI: https://developer.apple.com/documentation/corebluetooth/cbperipheral/1519111-readrssi
func (p Peripheral) ReadRSSI() {
	C.cb_prph_read_rssi(p.ptr)
}
