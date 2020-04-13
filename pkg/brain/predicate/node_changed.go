package predicate

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var nodeChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("nodeChanged")

var NodeChangedFuncs = predicate.Funcs{
	GenericFunc: func(e event.GenericEvent) bool {
		if e.Meta == nil {
			nodeChangedPredicateLog.Error(nil, "Received GenericEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			nodeChangedPredicateLog.Error(nil, "Received GenericEvent without runtime object", "event", e)
			return false
		}
		if object.IsNodeObject(e.Object) {
			// NB(thxCode) ignores all generic events of Node
			return false
		}
		return true
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.MetaOld == nil {
			nodeChangedPredicateLog.Error(nil, "Received UpdateEvent without old metadata", "event", e)
			return false
		}
		if e.ObjectOld == nil {
			nodeChangedPredicateLog.Error(nil, "Received UpdateEvent without old runtime object", "event", e)
			return false
		}
		if object.IsNodeObject(e.ObjectOld) {
			// NB(thxCode) ignores the update event of Node when:
			// - the node is existed
			if e.MetaOld.GetDeletionTimestamp().IsZero() {
				return false
			}
			nodeChangedPredicateLog.V(0).Info("Accept UpdateEvent", "key", object.GetNamespacedName(e.MetaOld))
			return true
		}
		return true
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Meta == nil {
			nodeChangedPredicateLog.Error(nil, "Received DeleteEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			nodeChangedPredicateLog.Error(nil, "Received DeleteEvent without runtime object", "event", e)
			return false
		}
		if object.IsNodeObject(e.Object) {
			// ignores the delete event of Node when:
			// - the node isn't existed
			if !e.Meta.GetDeletionTimestamp().IsZero() {
				return false
			}
			nodeChangedPredicateLog.V(0).Info("Accept DeleteEvent", "key", object.GetNamespacedName(e.Meta))
			return true
		}
		return true
	},
}
