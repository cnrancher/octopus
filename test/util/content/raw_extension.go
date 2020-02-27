package content

import (
	jsoniter "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/runtime"
)

func ToRawExtension(content interface{}) *runtime.RawExtension {
	if content == nil {
		return nil
	}
	switch t := content.(type) {
	case []byte:
		return &runtime.RawExtension{Raw: t}
	case string:
		return &runtime.RawExtension{Raw: []byte(t)}
	default:
		bs, _ := jsoniter.Marshal(content)
		return &runtime.RawExtension{Raw: bs}
	}
}
