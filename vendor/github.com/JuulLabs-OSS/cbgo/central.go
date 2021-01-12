package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import "unsafe"

// Central: https://developer.apple.com/documentation/corebluetooth/cbcentral
type Central struct {
	ptr unsafe.Pointer
}

// Identifier: https://developer.apple.com/documentation/corebluetooth/cbpeer/1620687-identifier
func (c Central) Identifier() UUID {
	cstr := C.cb_peer_identifier(c.ptr)
	return MustParseUUID(C.GoString(cstr))
}

// MaximumUpdateValueLength: https://developer.apple.com/documentation/corebluetooth/cbcentral/1408800-maximumupdatevaluelength
func (c Central) MaximumUpdateValueLength() int {
	return int(C.cb_cent_maximum_update_len(c.ptr))
}
