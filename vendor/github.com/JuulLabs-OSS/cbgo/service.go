package cbgo

import "unsafe"

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// Service: https://developer.apple.com/documentation/corebluetooth/cbservice
type Service struct {
	ptr unsafe.Pointer
}

// UUID: https://developer.apple.com/documentation/corebluetooth/cbattribute/1620638-uuid
func (s Service) UUID() UUID {
	cstr := C.cb_svc_uuid(s.ptr)
	return MustParseUUID(C.GoString(cstr))
}

// Peripheral: https://developer.apple.com/documentation/corebluetooth/cbservice/1434334-peripheral
func (s Service) Peripheral() Peripheral {
	prphPtr := C.cb_svc_peripheral(s.ptr)
	return Peripheral{prphPtr}
}

// IsPrimary: https://developer.apple.com/documentation/corebluetooth/cbservice/1434326-isprimary
func (s Service) IsPrimary() bool {
	return bool(C.cb_svc_is_primary(s.ptr))
}

// Characteristics: https://developer.apple.com/documentation/corebluetooth/cbservice/1434319-characteristics
func (s Service) Characteristics() []Characteristic {
	oa := C.cb_svc_characteristics(s.ptr)
	defer C.free(unsafe.Pointer(oa.objs))

	chrs := make([]Characteristic, oa.count)
	for i, _ := range chrs {
		obj := getObjArrElem(&oa, i)
		chrs[i] = Characteristic{ptr: obj}
	}

	return chrs
}

// IncludedServices: https://developer.apple.com/documentation/corebluetooth/cbservice/1434324-includedservices
func (s Service) IncludedServices() []Service {
	oa := C.cb_svc_included_svcs(s.ptr)
	defer C.free(unsafe.Pointer(oa.objs))

	svcs := make([]Service, oa.count)
	for i, _ := range svcs {
		obj := getObjArrElem(&oa, i)
		svcs[i] = Service{ptr: obj}
	}

	return svcs
}
