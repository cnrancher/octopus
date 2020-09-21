package cbgo

import "unsafe"

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// MutableCharacteristic: https://developer.apple.com/documentation/corebluetooth/cbmutablecharacteristic
type MutableCharacteristic struct {
	ptr unsafe.Pointer
}

func NewMutableCharacteristic(uuid UUID, properties CharacteristicProperties,
	value []byte, permissions AttributePermissions) MutableCharacteristic {

	cuuid := C.CString(uuid.String())
	defer C.free(unsafe.Pointer(cuuid))

	cvalue := byteSliceToByteArr(value)
	defer C.free(unsafe.Pointer(cvalue.data))

	return MutableCharacteristic{
		ptr: unsafe.Pointer(C.cb_mchr_alloc(cuuid, C.int(properties), &cvalue, C.int(permissions))),
	}
}

// Characteristic converts a MutableCharacteristic into its underlying
// Characteristic.
func (c MutableCharacteristic) Characteristic() Characteristic {
	return Characteristic{c.ptr}
}

// SetDescriptors: https://developer.apple.com/documentation/corebluetooth/cbmutablecharacteristic/1518827-descriptors
func (c MutableCharacteristic) SetDescriptors(mdscs []MutableDescriptor) {
	dscs := mallocObjArr(len(mdscs))
	defer C.free(unsafe.Pointer(dscs.objs))

	for i, mdsc := range mdscs {
		setObjArrElem(&dscs, i, mdsc.ptr)
	}

	C.cb_mchr_set_descriptors(c.ptr, &dscs)
}

// SetValue: https://developer.apple.com/documentation/corebluetooth/cbmutablecharacteristic/1519121-value
func (c MutableCharacteristic) SetValue(val []byte) {
	ba := byteSliceToByteArr(val)
	defer C.free(unsafe.Pointer(ba.data))

	C.cb_mchr_set_value(c.ptr, &ba)
}
