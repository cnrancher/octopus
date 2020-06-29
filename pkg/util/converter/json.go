package converter

import (
	stdjson "encoding/json"
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.Config{
	EscapeHTML:                    false,
	SortMapKeys:                   true,
	ValidateJsonRawMessage:        true,
	MarshalFloatWith6Digits:       true,
	ObjectFieldMustBeSimpleString: true,
}.Froze()

// UnmarshalJSON decodes the input JSON into the input handler.
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalJSON encodes the input object as JSON.
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// TryUnmarshalJSON is the same as UnmarshalJSON but doesn't return error.
func TryUnmarshalJSON(data []byte, v interface{}) {
	_ = UnmarshalJSON(data, v)
}

// TryMarshalJson is the same as MarshalJSON but doesn't return error.
func TryMarshalJSON(v interface{}) []byte {
	var bytes, _ = MarshalJSON(v)
	return bytes
}

func init() {
	// borrowed from https://github.com/json-iterator/go/issues/145#issuecomment-323483602
	decodeNumberAsInt64IfPossible := func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		switch iter.WhatIsNext() {
		case jsoniter.NumberValue:
			var number stdjson.Number
			iter.ReadVal(&number)
			i, err := strconv.ParseInt(string(number), 10, 64)
			if err == nil {
				*(*interface{})(ptr) = i
				return
			}
			f, err := strconv.ParseFloat(string(number), 64)
			if err == nil {
				*(*interface{})(ptr) = f
				return
			}
		default:
			*(*interface{})(ptr) = iter.Read()
		}
	}
	jsoniter.RegisterTypeDecoderFunc("interface {}", decodeNumberAsInt64IfPossible)
}
