package collection

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func StringMapCopy(source map[string]string) map[string]string {
	return StringMapCopyInto(source, make(map[string]string, len(source)))
}

func StringMapCopyInto(source, destination map[string]string) map[string]string {
	if destination == nil {
		return nil
	}
	if len(source) == 0 {
		return destination
	}

	for k, v := range source {
		destination[k] = v
	}
	return destination
}

func DiffStringMap(left, right map[string]string) bool {
	for lk, lv := range left {
		if rv, exist := right[lk]; !exist {
			return true
		} else if lv != rv {
			return true
		}
	}
	return false
}

func FormatStringMap(m map[string]string, splitter string) string {
	if splitter == "" {
		splitter = ","
	}

	var keySet = sets.NewString()
	for k := range m {
		keySet.Insert(k)
	}
	var keysLen = keySet.Len()
	var keys = keySet.List()

	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(`"`)
		builder.WriteString(m[k])
		builder.WriteString(`"`)

		if keysLen > 1 {
			builder.WriteString(splitter)
		}
		keysLen--
	}
	return builder.String()
}
