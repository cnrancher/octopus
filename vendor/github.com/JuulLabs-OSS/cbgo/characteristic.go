package cbgo

import "unsafe"

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// CharacteristicProperties: https://developer.apple.com/documentation/corebluetooth/cbcharacteristicproperties
type CharacteristicProperties int

const (
	CharacteristicPropertyBroadcast                  = CharacteristicProperties(C.CBCharacteristicPropertyBroadcast)
	CharacteristicPropertyRead                       = CharacteristicProperties(C.CBCharacteristicPropertyRead)
	CharacteristicPropertyWriteWithoutResponse       = CharacteristicProperties(C.CBCharacteristicPropertyWriteWithoutResponse)
	CharacteristicPropertyWrite                      = CharacteristicProperties(C.CBCharacteristicPropertyWrite)
	CharacteristicPropertyNotify                     = CharacteristicProperties(C.CBCharacteristicPropertyNotify)
	CharacteristicPropertyIndicate                   = CharacteristicProperties(C.CBCharacteristicPropertyIndicate)
	CharacteristicPropertyAuthenticatedSignedWrites  = CharacteristicProperties(C.CBCharacteristicPropertyAuthenticatedSignedWrites)
	CharacteristicPropertyExtendedProperties         = CharacteristicProperties(C.CBCharacteristicPropertyExtendedProperties)
	CharacteristicPropertyNotifyEncryptionRequired   = CharacteristicProperties(C.CBCharacteristicPropertyNotifyEncryptionRequired)
	CharacteristicPropertyIndicateEncryptionRequired = CharacteristicProperties(C.CBCharacteristicPropertyIndicateEncryptionRequired)
)

func chrWriteType(withRsp bool) C.int {
	if withRsp {
		return C.CBCharacteristicWriteWithResponse
	} else {
		return C.CBCharacteristicWriteWithoutResponse
	}
}

// Characteristic: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic
type Characteristic struct {
	ptr unsafe.Pointer
}

// UUID: https://developer.apple.com/documentation/corebluetooth/cbattribute/1620638-uuid
func (c Characteristic) UUID() UUID {
	cstr := C.cb_chr_uuid(c.ptr)
	return MustParseUUID(C.GoString(cstr))
}

// Service: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic/1518728-service
func (c Characteristic) Service() Service {
	ptr := C.cb_chr_service(c.ptr)
	return Service{ptr}
}

// Value: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic/1518878-value
func (c Characteristic) Value() []byte {
	ba := C.cb_chr_value(c.ptr)
	return byteArrToByteSlice(&ba)
}

// Descriptors: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic/1518957-descriptors
func (c Characteristic) Descriptors() []Descriptor {
	oa := C.cb_chr_descriptors(c.ptr)
	defer C.free(unsafe.Pointer(oa.objs))

	dscs := make([]Descriptor, oa.count)
	for i, _ := range dscs {
		obj := getObjArrElem(&oa, i)
		dscs[i] = Descriptor{ptr: obj}
	}

	return dscs
}

// Properties: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic/1519010-properties
func (c Characteristic) Properties() CharacteristicProperties {
	return CharacteristicProperties(C.cb_chr_properties(c.ptr))
}

// IsNotifying: https://developer.apple.com/documentation/corebluetooth/cbcharacteristic/1519057-isnotifying
func (c Characteristic) IsNotifying() bool {
	return bool(C.cb_chr_is_notifying(c.ptr))
}
