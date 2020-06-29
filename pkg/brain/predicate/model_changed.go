package predicate

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var modelChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("modelChanged")

type ModelChangedPredicate struct {
	predicate.Funcs
}

func (ModelChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld == nil || e.MetaNew == nil || e.ObjectNew == nil || e.ObjectOld == nil {
		return false
	}

	// doesn't handle non-CRD object
	if !object.IsCustomResourceDefinitionObject(e.ObjectOld) {
		return true
	}

	var crdOld = object.ToCustomResourceDefinitionObject(e.ObjectOld)
	var crdNew = object.ToCustomResourceDefinitionObject(e.ObjectNew)

	// handles when it's not backward compatible
	if !isBackwardCompatibleCRDVersions(crdOld.Spec.Versions, crdNew.Spec.Versions) {
		modelChangedPredicateLog.V(5).Info("Accept UpdateEvent as bad backward compatible", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}

	return false
}

// isBackwardCompatibleCRDVersions detects if the new stored versions compatible with those old ones.
func isBackwardCompatibleCRDVersions(oldStoredVersions, newStoredVersions []apiextensionsv1.CustomResourceDefinitionVersion) bool {
	if len(oldStoredVersions) == 0 || len(newStoredVersions) == 0 {
		return false
	}

	var newStoredVersionsIndex = make(map[string]struct{}, len(newStoredVersions))
	for _, newStoredVersion := range newStoredVersions {
		if newStoredVersion.Served {
			newStoredVersionsIndex[newStoredVersion.Name] = struct{}{}
		}
	}
	for _, oldStoredVersion := range oldStoredVersions {
		if _, exist := newStoredVersionsIndex[oldStoredVersion.Name]; !exist {
			return false
		}
	}
	return true
}
