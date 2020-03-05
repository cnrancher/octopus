package predicate

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var modelChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("ModelChanged")

var ModelChangedFuncs = predicate.Funcs{
	GenericFunc: func(e event.GenericEvent) bool {
		if e.Meta == nil {
			modelChangedPredicateLog.Error(nil, "received GenericEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			modelChangedPredicateLog.Error(nil, "received GenericEvent without runtime object", "event", e)
			return false
		}
		if object.IsCustomResourceDefinitionObject(e.Object) {
			// NB(thxCode) ignores all generic events of CRD
			return false
		}
		return true
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.MetaOld == nil {
			modelChangedPredicateLog.Error(nil, "received UpdateEvent without old metadata", "event", e)
			return false
		}
		if e.ObjectOld == nil {
			modelChangedPredicateLog.Error(nil, "received UpdateEvent without old runtime object", "event", e)
			return false
		}
		if object.IsCustomResourceDefinitionObject(e.ObjectOld) {
			// NB(thxCode) ignores all generic events of CRD when:
			// - the CRD is existed
			// - TODO verify version
			if e.MetaOld.GetDeletionTimestamp().IsZero() {
				return false
			}
			modelChangedPredicateLog.V(0).Info("accept UpdateEvent", "key", object.GetNamespacedName(e.MetaOld))
			return true
		}
		return true
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Meta == nil {
			modelChangedPredicateLog.Error(nil, "received DeleteEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			modelChangedPredicateLog.Error(nil, "received DeleteEvent without runtime object", "event", e)
			return false
		}
		if object.IsCustomResourceDefinitionObject(e.Object) {
			// NB(thxCode) ignores the delete event of CRD when:
			// - the CRD isn't existed
			if !e.Meta.GetDeletionTimestamp().IsZero() {
				return false
			}
			modelChangedPredicateLog.V(0).Info("accept DeleteEvent", "key", object.GetNamespacedName(e.Meta))
			return true
		}
		return true
	},
}
