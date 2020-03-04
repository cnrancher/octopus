package model

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetCRDNameOfGroupVersionKind(gvk schema.GroupVersionKind) string {
	var gk = gvk.GroupKind()
	if gk.Kind == "" || gk.Group == "" {
		return ""
	}
	return strings.ToLower(fmt.Sprintf("%ss.%s", gk.Kind, gk.Group))
}
