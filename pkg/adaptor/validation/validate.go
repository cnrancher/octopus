package validation

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	IsQualifiedName = validation.IsQualifiedName
)

func IsSocketFile(filename string) bool {
	return strings.HasSuffix(filename, ".socket")
}
