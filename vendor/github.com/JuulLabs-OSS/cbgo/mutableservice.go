package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import "unsafe"

// MutableService: https://developer.apple.com/documentation/corebluetooth/cbmutableservice
type MutableService struct {
	ptr unsafe.Pointer
}

func NewMutableService(uuid UUID, primary bool) MutableService {
	cuuid := C.CString(uuid.String())
	defer C.free(unsafe.Pointer(cuuid))

	return MutableService{
		ptr: unsafe.Pointer(C.cb_msvc_alloc(cuuid, C.bool(primary))),
	}
}

// Service converts a MutableService into its underlying Service.
func (s MutableService) Service() Service {
	return Service{s.ptr}
}

// SetCharacteristics: https://developer.apple.com/documentation/corebluetooth/cbmutableservice/1434317-characteristics
func (s MutableService) SetCharacteristics(mchrs []MutableCharacteristic) {
	chrs := mallocObjArr(len(mchrs))
	defer C.free(unsafe.Pointer(chrs.objs))

	for i, mchr := range mchrs {
		setObjArrElem(&chrs, i, mchr.ptr)
	}

	C.cb_msvc_set_characteristics(s.ptr, &chrs)
}

// SetIncludedServices: https://developer.apple.com/documentation/corebluetooth/cbmutableservice/1434320-includedservices
func (s MutableService) SetIncludedServices(msvcs []MutableService) {
	svcs := mallocObjArr(len(msvcs))
	defer C.free(unsafe.Pointer(svcs.objs))

	for i, msvc := range msvcs {
		setObjArrElem(&svcs, i, msvc.ptr)
	}

	C.cb_msvc_set_included_services(s.ptr, &svcs)
}
