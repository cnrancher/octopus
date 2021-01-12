package cbgo

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

const UUID16StringLength = 4
const UUID128StringLength = 36

// UUID is a 16-bit or 128-bit UUID.
type UUID []byte

func reverse(bs []byte) []byte {
	c := make([]byte, len(bs))
	for i, b := range bs {
		c[len(c)-1-i] = b
	}

	return c
}

// UUID16 constructs a 16-bit UUID from a uint16.
func UUID16(i uint16) UUID {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return UUID(b)
}

// UUID16 constructs a 128-bit UUID from a 16-element byte slice.
func UUID128(b []byte) (UUID, error) {
	if len(b) != 16 {
		return nil, fmt.Errorf("failed to construct UUID128: wrong length: have=%d want=16", len(b))
	}

	return UUID(b), nil
}

// ParseUUID16 parses a UUID string with the form:
//     1234
func ParseUUID16(s string) (UUID, error) {
	if len(s) != UUID16StringLength {
		return nil, fmt.Errorf("invalid UUID16: %s", s)
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invald UUID16: %v", err)
	}

	return UUID(reverse(b)), nil
}

// ParseUUID128 parses a UUID string with the form:
//     01234567-89ab-cdef-0123-456789abcdef
func ParseUUID128(s string) (UUID, error) {
	if len(s) != UUID128StringLength {
		return nil, fmt.Errorf("invalid UUID128: %s", s)
	}

	b := make([]byte, 16)

	off := 0
	for i := 0; i < 36; {
		switch i {
		case 8, 13, 18, 23:
			if s[i] != '-' {
				return nil, fmt.Errorf("invalid UUID128: %s", s)
			}
			i++

		default:
			u64, err := strconv.ParseUint(s[i:i+2], 16, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid UUID128: %s", s)
			}
			b[off] = byte(u64)
			i += 2
			off++
		}
	}

	return UUID(reverse(b)), nil
}

// ParseUUID parses a string representing a 16-bit or 128-bit UUID.
func ParseUUID(s string) (UUID, error) {
	switch len(s) {
	case UUID16StringLength:
		return ParseUUID16(s)
	case UUID128StringLength:
		return ParseUUID128(s)
	default:
		return nil, fmt.Errorf("invalid UUID string: %s", s)
	}
}

// MustParseUUID is like ParseUUID except it panics on failure.
func MustParseUUID(s string) UUID {
	uuid, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}

	return uuid
}

// String retruns a CoreBluetooth-friendly string representation of a UUID.
func (u UUID) String() string {
	b := reverse(u)

	switch len(b) {
	case 2:
		return fmt.Sprintf("%x", []byte(b))

	case 16:
		s := fmt.Sprintf("%x", []byte(b))
		// 01234567-89ab-cdef-0123-456789abcdef
		return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]

	default:
		btlog.Errorf("invalid UUID: %s", hex.EncodeToString(u))
		return ""
	}
}

// NSError: https://developer.apple.com/documentation/foundation/nserror
type NSError struct {
	msg  string
	code int
}

func (e *NSError) Message() string {
	return e.msg
}
func (e *NSError) Code() int {
	return e.code
}
func (e *NSError) Error() string {
	return e.msg
}
