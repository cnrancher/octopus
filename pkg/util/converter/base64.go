package converter

import (
	"encoding/base64"
	"strings"
)

// DecodeBase64String decodes the input string, which can
// accept padded or none padded format.
func DecodeBase64String(str string) ([]byte, error) {
	// normalizes to std encoding format
	str = strings.ReplaceAll(str, "-", "+")
	str = strings.ReplaceAll(str, "_", "/")
	// normalizes to no padding format
	str = strings.TrimRight(str, "=")
	var b, err = base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// DecodeBase64 decodes the input bytes, which can
// accept padded or none padded format.
func DecodeBase64(src []byte) ([]byte, error) {
	return DecodeBase64String(UnsafeBytesToString(src))
}

// EncodeBase64 encodes the input bytes,
// and then output standard format.
func EncodeBase64(src []byte) []byte {
	var enc = base64.StdEncoding
	var ret = make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(ret, src)
	return ret
}

// EncodeBase64ToString encodes the input bytes,
// and then output standard format string.
func EncodeBase64ToString(src []byte) string {
	return UnsafeBytesToString(EncodeBase64(src))
}
