package cbgo

import "unsafe"

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// MutableDescriptor: https://developer.apple.com/documentation/corebluetooth/cbmutabledescriptor
type MutableDescriptor struct {
	ptr unsafe.Pointer
}

func NewMutableDescriptor(uuid UUID, value []byte) MutableDescriptor {
	cuuid := C.CString(uuid.String())
	defer C.free(unsafe.Pointer(cuuid))

	cvalue := byteSliceToByteArr(value)
	defer C.free(unsafe.Pointer(cvalue.data))

	return MutableDescriptor{
		ptr: unsafe.Pointer(C.cb_mdsc_alloc(cuuid, &cvalue)),
	}
}

// Descriptor converts a MutableDescriptor into its underlying Descriptor.
func (d MutableDescriptor) Descriptor() Descriptor {
	return Descriptor{d.ptr}
}
