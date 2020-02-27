package predicate

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var nodeChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("NodeChanged")

var NodeChangedFuncs = predicate.Funcs{
	GenericFunc: func(e event.GenericEvent) bool {
		if e.Meta == nil {
			nodeChangedPredicateLog.Error(nil, "received GenericEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			nodeChangedPredicateLog.Error(nil, "received GenericEvent without runtime object", "event", e)
			return false
		}
		// ignores all generic events of Node
		if object.IsNodeObject(e.Object) {
			nodeChangedPredicateLog.V(0).Info("ignore GenericEvent")
			return false
		}
		return true
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.MetaOld == nil {
			nodeChangedPredicateLog.Error(nil, "received UpdateEvent without old metadata", "event", e)
			return false
		}
		if e.ObjectOld == nil {
			nodeChangedPredicateLog.Error(nil, "received UpdateEvent without old runtime object", "event", e)
			return false
		}
		// ignores the update event of Node when:
		// - the node is existed
		if object.IsNodeObject(e.ObjectOld) {
			if e.MetaOld.GetDeletionTimestamp().IsZero() {
				nodeChangedPredicateLog.V(0).Info("ignore UpdateEvent")
				return false
			}
		}
		return true
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		if e.Meta == nil {
			nodeChangedPredicateLog.Error(nil, "received DeleteEvent without metadata", "event", e)
			return false
		}
		if e.Object == nil {
			nodeChangedPredicateLog.Error(nil, "received DeleteEvent without runtime object", "event", e)
			return false
		}
		// ignores the delete event of Node when:
		// - the node isn't existed
		if object.IsNodeObject(e.Object) {
			if !e.Meta.GetDeletionTimestamp().IsZero() {
				nodeChangedPredicateLog.V(0).Info("ignore DeleteEvent")
				return false
			}
		}
		return true
	},
}
