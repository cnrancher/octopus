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
		// ignores all generic events of CRD
		if object.IsCustomResourceDefinitionObject(e.Object) {
			modelChangedPredicateLog.V(0).Info("ignore GenericEvent")
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
		// ignores all generic events of CRD when:
		// - the CRD is existed and ... TODO verify version
		if object.IsCustomResourceDefinitionObject(e.ObjectOld) {
			if e.MetaOld.GetDeletionTimestamp().IsZero() {
				// TODO verify version
				modelChangedPredicateLog.V(0).Info("ignore UpdateEvent")
				return false
			}
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
		// ignores the delete event of CRD when:
		// - the CRD isn't existed
		if object.IsCustomResourceDefinitionObject(e.Object) {
			if !e.Meta.GetDeletionTimestamp().IsZero() {
				modelChangedPredicateLog.V(0).Info("ignore DeleteEvent")
				return false
			}
		}
		return true
	},
}
