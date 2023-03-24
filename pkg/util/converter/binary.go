package converter

import (
	"errors"
	"strings"
)

// DecodeBinaryString converts the binary string value to a byte array.
func DecodeBinaryString(str string) ([]byte, error) {
	// the length of the string must be a multiple of 8.
	if len(str)%8 != 0 {
		return nil, errors.New("the length of binary string value must be a multiple of eight")
	}
	// for example, when input "1100110100000001", every 8 chars are grouped,
	// and then the 8-chars string in a group is converted into one byte.
	// binary value  :  1  1  0  0  1  1  0  1  0  0  0  0  0  0  0  1|  <-- input
	// byte value    :  *  *  *  *  *  *  *205| *  *  *  *  *  *  *  1|  <-- output
	var val []byte
	for si := 0; si < len(str); si = si + 8 {
		var v byte
		for sj := si + 7; sj >= si; sj-- {
			if str[sj] == '1' {
				v = v | (0b1 << (7 - sj&7))
			}
		}
		val = append(val, v)
	}
	return val, nil
}

// EncodeBinaryToString parses the binary string value from a byte array.
func EncodeBinaryToString(data []byte) string {
	// for example, when input [205,1], every byte  are grouped,
	// every byte is converted into a 8-chars string composed of 0 and 1.
	// byte value    :  *  *  *  *  *  *  *205| *  *  *  *  *  *  *  1|  <-- input
	// binary value  :  1  1  0  0  1  1  0  1  0  0  0  0  0  0  0  1|  <-- output
	var valSb = &strings.Builder{}
	for _, b := range data {
		for bi := 7; bi >= 0; bi-- {
			if (b>>bi)&0b1 == 0b1 {
				valSb.WriteString("1")
			} else {
				valSb.WriteString("0")
			}
		}
	}
	return valSb.String()
}
