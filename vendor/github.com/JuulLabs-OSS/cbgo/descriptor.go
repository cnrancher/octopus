package cbgo

import "unsafe"

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// Descriptor: https://developer.apple.com/documentation/corebluetooth/cbdescriptor
type Descriptor struct {
	ptr unsafe.Pointer
}

// UUID: https://developer.apple.com/documentation/corebluetooth/cbattribute/1620638-uuid
func (d Descriptor) UUID() UUID {
	cstr := C.cb_dsc_uuid(d.ptr)
	return MustParseUUID(C.GoString(cstr))
}

// Characteristic: https://developer.apple.com/documentation/corebluetooth/cbdescriptor/1519035-characteristic
func (d Descriptor) Characteristic() Characteristic {
	ptr := C.cb_dsc_characteristic(d.ptr)
	return Characteristic{ptr}
}

// Value: https://developer.apple.com/documentation/corebluetooth/cbdescriptor/1518778-value
func (d Descriptor) Value() []byte {
	ba := C.cb_dsc_value(d.ptr)
	return byteArrToByteSlice(&ba)
}
