package converter

import (
	"reflect"
	"unsafe"
)

func UnsafeBytesToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func UnsafeStringToBytes(s string) (bytes []byte) {
	var slice = (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	var str = (*reflect.StringHeader)(unsafe.Pointer(&s))
	slice.Len = str.Len
	slice.Cap = str.Len
	slice.Data = str.Data
	return bytes
}
