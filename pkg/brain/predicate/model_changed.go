package predicate

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var modelChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("modelChanged")

var ModelChangedFuncs = predicate.Funcs{
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Meta == nil {
			modelChangedPredicateLog.Error(nil, "Received DeleteEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			modelChangedPredicateLog.Error(nil, "Received DeleteEvent without runtime object", "event", e)
			return false
		}
		if object.IsCustomResourceDefinitionObject(e.Object) {
			// NB(thxCode) ignores the delete event of CRD when:
			// - the CRD isn't existed
			if !e.Meta.GetDeletionTimestamp().IsZero() {
				return false
			}
			modelChangedPredicateLog.V(0).Info("Accept DeleteEvent", "key", object.GetNamespacedName(e.Meta))
			return true
		}
		return true
	},
}
