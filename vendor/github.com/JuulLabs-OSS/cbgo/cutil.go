package cbgo

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreBluetooth

#import "bt.h"
*/
import "C"

import (
	"unsafe"
)

func mallocArr(numElems int, elemSize uintptr) unsafe.Pointer {
	return C.malloc(C.size_t(numElems) * C.size_t(elemSize))
}

// getArrElemAddr retrieves the address of an element from a C array.
func getArrElemAddr(arr unsafe.Pointer, elemSize uintptr, idx int) unsafe.Pointer {
	base := uintptr(arr)
	off := uintptr(idx) * elemSize
	cur := base + off
	return unsafe.Pointer(cur)
}

// freeArrElems frees each element of a C-array of pointers.
func freeArrElems(arr unsafe.Pointer, elemSize uintptr, count int) {
	for i := 0; i < count; i++ {
		base := uintptr(arr)
		off := uintptr(i) * elemSize
		addr := base + off
		pp := unsafe.Pointer(addr)
		ptr := *(*unsafe.Pointer)(pp)
		C.free(ptr)
	}
}

func byteArrToByteSlice(byteArr *C.struct_byte_arr) []byte {
	return C.GoBytes(unsafe.Pointer(byteArr.data), byteArr.length)
}

func byteSliceToByteArr(b []byte) C.struct_byte_arr {
	if len(b) == 0 {
		return C.struct_byte_arr{
			data:   nil,
			length: 0,
		}
	} else {
		return C.struct_byte_arr{
			data:   (*C.uint8_t)(C.CBytes(b)),
			length: C.int(len(b)),
		}
	}
}

func stringArrToStringSlice(sa *C.struct_string_arr) []string {
	ss := make([]string, sa.count)
	for i, _ := range ss {
		ss[i] = getStrArrElem(sa, i)
	}

	return ss
}

func stringSliceToArr(ss []string) C.struct_string_arr {
	if len(ss) == 0 {
		return C.struct_string_arr{}
	}

	ptr := mallocArr(len(ss), unsafe.Sizeof((*C.char)(nil)))

	carr := (*[1<<30 - 1]*C.char)(ptr)
	for i, s := range ss {
		carr[i] = C.CString(s)
	}

	return C.struct_string_arr{
		strings: (**C.char)(ptr),
		count:   C.int(len(ss)),
	}
}

// getStrArrElem retrieves an element from a `struct string_arr` C object.
func getStrArrElem(sa *C.struct_string_arr, idx int) string {
	elemSize := unsafe.Sizeof(*sa.strings)

	ptr := getArrElemAddr(unsafe.Pointer(sa.strings), elemSize, idx)
	cstr := *(**C.char)(ptr)
	return C.GoString(cstr)
}

func uuidsToStrArr(uuids []UUID) C.struct_string_arr {
	var ss []string
	for _, u := range uuids {
		ss = append(ss, u.String())
	}

	return stringSliceToArr(ss)
}

func strArrToUUIDs(sa *C.struct_string_arr) ([]UUID, error) {
	if sa == nil || sa.count == 0 {
		return nil, nil
	}

	uuids := make([]UUID, sa.count)

	ss := stringArrToStringSlice(sa)
	for i, s := range ss {
		uuid, err := ParseUUID(s)
		if err != nil {
			return nil, err
		}
		uuids[i] = uuid
	}

	return uuids, nil
}

func mustStrArrToUUIDs(sa *C.struct_string_arr) []UUID {
	uuids, err := strArrToUUIDs(sa)
	if err != nil {
		panic(err)
	}

	return uuids
}

func freeStrArr(sa *C.struct_string_arr) {
	freeArrElems(unsafe.Pointer(sa.strings), unsafe.Sizeof(*sa.strings), int(sa.count))
	C.free(unsafe.Pointer(sa.strings))
}

func btErrorToNSError(e *C.struct_bt_error) error {
	if e == nil || e.msg == nil {
		return nil
	} else {
		return &NSError{
			msg:  C.GoString(e.msg),
			code: int(e.code),
		}
	}
}

func mallocObjArr(count int) C.struct_obj_arr {
	if count <= 0 {
		return C.struct_obj_arr{
			objs:  nil,
			count: 0,
		}
	}

	data := mallocArr(count, unsafe.Sizeof(unsafe.Pointer(nil)))
	return C.struct_obj_arr{
		objs:  (*unsafe.Pointer)(data),
		count: C.int(count),
	}
}

// getStrArrElem retrieves an element from a `struct obj_arr` C object.
func getObjArrElem(oa *C.struct_obj_arr, idx int) unsafe.Pointer {
	addr := getArrElemAddr(unsafe.Pointer(oa.objs), unsafe.Sizeof(*oa.objs), idx)
	dptr := (*unsafe.Pointer)(addr)
	return *dptr
}

// setStrArrElem assigns an element in a `struct obj_arr` C object.
func setObjArrElem(oa *C.struct_obj_arr, idx int, val unsafe.Pointer) {
	addr := getArrElemAddr(unsafe.Pointer(oa.objs), unsafe.Sizeof(*oa.objs), idx)
	dptr := (*unsafe.Pointer)(addr)
	*dptr = val
}
