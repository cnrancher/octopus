package uuid

import (
	"github.com/rancher/octopus/pkg/util/converter"
)

const uuidLength = 36
const uuidWithoutHyphenLength = 32

func Truncate(uuid string, remainingLength int) string {
	return ring(uuid).truncate(0, uuidWithoutHyphenLength-remainingLength)
}

const hyphen = byte('-')

var offsets = map[byte]int{
	byte('0'): 1,
	byte('1'): 2,
	byte('2'): 3,
	byte('3'): 4,
	byte('4'): 5,
	byte('5'): 6,
	byte('6'): 7,
	byte('7'): 8,
	byte('8'): 9,
	byte('9'): 10,
	byte('a'): 11,
	byte('b'): 12,
	byte('c'): 13,
	byte('d'): 14,
	byte('e'): 15,
	byte('f'): 16,
	byte('A'): 11,
	byte('B'): 12,
	byte('C'): 13,
	byte('D'): 14,
	byte('E'): 15,
	byte('F'): 16,
}

type ring []byte

func (r ring) truncate(start, truncateLength int) string {
	if truncateLength == uuidWithoutHyphenLength {
		return ""
	}

	if len(r) <= truncateLength || truncateLength < 0 {
		truncateLength = 0
	}
	for i := 0; i < truncateLength; i++ {
		start = r.next(start)
	}

	return r.toString(uuidWithoutHyphenLength - truncateLength)
}

func (r ring) toString(remainingLength int) string {
	var dst = make([]byte, 0, remainingLength)
	for _, b := range r {
		if b != hyphen {
			dst = append(dst, b)
		}
	}
	return converter.UnsafeBytesToString(dst)
}

func (r ring) next(start int) int {
	for {
		var b = r[start]
		if b != hyphen {
			r[start] = hyphen
			start = r.ranging(start, offsets[b])
			break
		}

		start = r.ranging(start, 1)
	}

	return start
}

func (r ring) ranging(start, offset int) int {
	start += offset
	start = start % uuidLength
	return start
}
