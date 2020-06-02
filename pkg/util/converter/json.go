package converter

import jsoniter "github.com/json-iterator/go"

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
