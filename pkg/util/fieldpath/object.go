package fieldpath

// Borrowed from k8s.io/kubernetes/pkg/fieldpath/fieldpath.go

import (
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/converter"
)

// ExtractObjectFieldPathAsBytes is extracts the field from the given Object
// and returns it as a byte array.
func ExtractObjectFieldPathAsBytes(obj runtime.Object, fieldPath string) ([]byte, error) {
	var accessor, err = meta.Accessor(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access obj: %v", obj)
	}

	var str string
	switch fieldPath {
	case "metadata.annotations":
		str = collection.FormatStringMap(accessor.GetAnnotations(), ";")
	case "metadata.labels":
		str = collection.FormatStringMap(accessor.GetLabels(), ";")
	case "metadata.name":
		str = accessor.GetName()
	case "metadata.namespace":
		str = accessor.GetNamespace()
	case "metadata.uid":
		str = string(accessor.GetUID())
	default:
		var path, subscript, ok = splitMaybeSubscriptedPath(fieldPath)
		if !ok {
			return nil, errors.Errorf("unsupported fieldPath: %s", fieldPath)
		}
		switch path {
		case "metadata.annotations":
			if errs := validation.IsQualifiedName(strings.ToLower(subscript)); len(errs) != 0 {
				return nil, errors.Errorf("invalid key subscript in %s: %s", fieldPath, strings.Join(errs, ";"))
			}
			str = accessor.GetAnnotations()[subscript]
		case "metadata.labels":
			if errs := validation.IsQualifiedName(subscript); len(errs) != 0 {
				return nil, errors.Errorf("invalid key subscript in %s: %s", fieldPath, strings.Join(errs, ";"))
			}
			str = accessor.GetLabels()[subscript]
		default:
			return nil, errors.Errorf("fieldPath %s doesn't support subscript", fieldPath)
		}
	}
	return converter.UnsafeStringToBytes(str), nil
}

// splitMaybeSubscriptedPath checks whether the specified fieldPath is
// subscripted, and
//  - if yes, this function splits the fieldPath into path and subscript, and
//    returns (path, subscript, true).
//  - if no, this function returns (fieldPath, "", false).
//
// Example inputs and outputs:
//  - "metadata.annotations['myKey']" --> ("metadata.annotations", "myKey", true)
//  - "metadata.annotations['a[b]c']" --> ("metadata.annotations", "a[b]c", true)
//  - "metadata.labels['']"           --> ("metadata.labels", "", true)
//  - "metadata.labels"               --> ("metadata.labels", "", false)
func splitMaybeSubscriptedPath(fieldPath string) (string, string, bool) {
	if !strings.HasSuffix(fieldPath, "']") {
		return fieldPath, "", false
	}
	s := strings.TrimSuffix(fieldPath, "']")
	parts := strings.SplitN(s, "['", 2)
	if len(parts) < 2 {
		return fieldPath, "", false
	}
	if len(parts[0]) == 0 {
		return fieldPath, "", false
	}
	return parts[0], parts[1], true
}
