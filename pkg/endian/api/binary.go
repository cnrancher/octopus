package api

// LittleEndianSwap is the swapping little-endian implementation of binary.ByteOrder.
var LittleEndianSwap littleEndianSwap

// BigEndianSwap is the swapping big-endian implementation of binary.ByteOrder.
var BigEndianSwap bigEndianSwap

type littleEndianSwap struct{}

func (littleEndianSwap) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[1]) | uint16(b[0])<<8
}

func (littleEndianSwap) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func (littleEndianSwap) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[1]) | uint32(b[0])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24
}

func (littleEndianSwap) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
}

func (littleEndianSwap) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[1]) | uint64(b[0])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 |
		uint64(b[5])<<32 | uint64(b[4])<<40 | uint64(b[7])<<48 | uint64(b[6])<<56
}

func (littleEndianSwap) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
	b[4] = byte(v >> 40)
	b[5] = byte(v >> 32)
	b[6] = byte(v >> 56)
	b[7] = byte(v >> 48)
}

func (littleEndianSwap) String() string { return "LittleEndianSwap" }

type bigEndianSwap struct{}

func (bigEndianSwap) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

func (bigEndianSwap) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func (bigEndianSwap) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[2]) | uint32(b[3])<<8 | uint32(b[0])<<16 | uint32(b[1])<<24
}

func (bigEndianSwap) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 24)
	b[2] = byte(v)
	b[3] = byte(v >> 8)
}

func (bigEndianSwap) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[6]) | uint64(b[7])<<8 | uint64(b[4])<<16 | uint64(b[5])<<24 |
		uint64(b[2])<<32 | uint64(b[3])<<40 | uint64(b[0])<<48 | uint64(b[1])<<56
}

func (bigEndianSwap) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 48)
	b[1] = byte(v >> 56)
	b[2] = byte(v >> 32)
	b[3] = byte(v >> 40)
	b[4] = byte(v >> 16)
	b[5] = byte(v >> 24)
	b[6] = byte(v)
	b[7] = byte(v >> 8)
}

func (bigEndianSwap) String() string { return "BigEndianSwap" }
