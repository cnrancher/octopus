package validation

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

func IsQualifiedName(name string) bool {
	return len(validation.IsQualifiedName(name)) == 0
}

func IsSocketFile(filename string) bool {
	return strings.HasSuffix(filename, api.SocketSuffix)
}

func IsSupportedVersion(version string) bool {
	for _, v := range api.SupportedVersions {
		if version == v {
			return true
		}
	}
	return false
}
