package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"
import "unsafe"

// AttributePermissions: https://developer.apple.com/documentation/corebluetooth/cbattributepermissions
type AttributePermissions int

const (
	AttributePermissionsReadable                = AttributePermissions(C.CBAttributePermissionsReadable)
	AttributePermissionsWriteable               = AttributePermissions(C.CBAttributePermissionsWriteable)
	AttributePermissionsReadEncryptionRequired  = AttributePermissions(C.CBAttributePermissionsReadEncryptionRequired)
	AttributePermissionsWriteEncryptionRequired = AttributePermissions(C.CBAttributePermissionsWriteEncryptionRequired)
)

// ATTError: https://developer.apple.com/documentation/corebluetooth/cbatterror
type ATTError int

const (
	ATTErrorSuccess                       = ATTError(C.CBATTErrorSuccess)
	ATTErrorInvalidHandle                 = ATTError(C.CBATTErrorInvalidHandle)
	ATTErrorReadNotPermitted              = ATTError(C.CBATTErrorReadNotPermitted)
	ATTErrorWriteNotPermitted             = ATTError(C.CBATTErrorWriteNotPermitted)
	ATTErrorInvalidPdu                    = ATTError(C.CBATTErrorInvalidPdu)
	ATTErrorInsufficientAuthentication    = ATTError(C.CBATTErrorInsufficientAuthentication)
	ATTErrorRequestNotSupported           = ATTError(C.CBATTErrorRequestNotSupported)
	ATTErrorInvalidOffset                 = ATTError(C.CBATTErrorInvalidOffset)
	ATTErrorInsufficientAuthorization     = ATTError(C.CBATTErrorInsufficientAuthorization)
	ATTErrorPrepareQueueFull              = ATTError(C.CBATTErrorPrepareQueueFull)
	ATTErrorAttributeNotFound             = ATTError(C.CBATTErrorAttributeNotFound)
	ATTErrorAttributeNotLong              = ATTError(C.CBATTErrorAttributeNotLong)
	ATTErrorInsufficientEncryptionKeySize = ATTError(C.CBATTErrorInsufficientEncryptionKeySize)
	ATTErrorInvalidAttributeValueLength   = ATTError(C.CBATTErrorInvalidAttributeValueLength)
	ATTErrorUnlikelyError                 = ATTError(C.CBATTErrorUnlikelyError)
	ATTErrorInsufficientEncryption        = ATTError(C.CBATTErrorInsufficientEncryption)
	ATTErrorUnsupportedGroupType          = ATTError(C.CBATTErrorUnsupportedGroupType)
	ATTErrorInsufficientResources         = ATTError(C.CBATTErrorInsufficientResources)
)

// ATTRequest: https://developer.apple.com/documentation/corebluetooth/cbattrequest
type ATTRequest struct {
	ptr unsafe.Pointer
}

// Central: https://developer.apple.com/documentation/corebluetooth/cbattrequest/1518995-central
func (r ATTRequest) Central() Central {
	ptr := C.cb_atr_central(r.ptr)
	return Central{unsafe.Pointer(ptr)}
}

// Characteristic: https://developer.apple.com/documentation/corebluetooth/cbattrequest/1518716-characteristic
func (r ATTRequest) Characteristic() Characteristic {
	ptr := C.cb_atr_characteristic(r.ptr)
	return Characteristic{unsafe.Pointer(ptr)}
}

// Value: https://developer.apple.com/documentation/corebluetooth/cbattrequest/1518795-value
func (r ATTRequest) Value() []byte {
	ba := C.cb_atr_value(r.ptr)
	return byteArrToByteSlice(&ba)
}

// SetValue: https://developer.apple.com/documentation/corebluetooth/cbattrequest/1518795-value
func (r ATTRequest) SetValue(v []byte) {
	ba := byteSliceToByteArr(v)
	defer C.free(unsafe.Pointer(ba.data))

	C.cb_atr_set_value(r.ptr, &ba)
}

// Offset: https://developer.apple.com/documentation/corebluetooth/cbattrequest/1518857-offset
func (r ATTRequest) Offset() int {
	return int(C.cb_atr_offset(r.ptr))
}
