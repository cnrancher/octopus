package predicate

import (
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var nodeChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("nodeChanged")

type NodeChangedPredicate struct {
	predicate.Funcs
}

func (NodeChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld == nil || e.MetaNew == nil || e.ObjectNew == nil || e.ObjectOld == nil {
		return false
	}

	// doesn't handle non-Node object
	if !object.IsNodeObject(e.ObjectOld) {
		return true
	}

	// handles when deleting
	if object.IsDeleted(e.MetaNew) {
		nodeChangedPredicateLog.V(5).Info("Accept UpdateEvent as deleted object", "key", object.GetNamespacedName(e.MetaOld))
		return true
	}

	var nodeOld = object.ToNodeObject(e.ObjectOld)
	var nodeNew = object.ToNodeObject(e.ObjectNew)

	// handles when changing addresses
	if diffNodeAddresses(nodeOld.Status.Addresses, nodeNew.Status.Addresses) {
		nodeChangedPredicateLog.V(5).Info("Accept UpdateEvent as diffed addresses", "key", object.GetNamespacedName(e.MetaOld))
		return true
	}

	// handles when changing nodes
	if diffNodeConditions(nodeOld.Status.Conditions, nodeNew.Status.Conditions) {
		nodeChangedPredicateLog.V(5).Info("Accept UpdateEvent as diffed conditions", "key", object.GetNamespacedName(e.MetaOld))
		return true
	}

	return false
}

func (NodeChangedPredicate) Delete(e event.DeleteEvent) bool {
	if e.Meta == nil || e.Object == nil {
		return false
	}

	// doesn't handle non-Node object
	if !object.IsNodeObject(e.Object) {
		return true
	}

	// NB(thxCode) there is a finalizer to handler the Node deletion event,
	// so with the finalizer, the deletion event can be changed to an update event.
	return false
}

// diffNodeAddresses compares the differences between two Node object's addresses.
// If there is a difference, it returns true.
func diffNodeAddresses(oldAddresses, newAddresses []corev1.NodeAddress) bool {
	if len(oldAddresses) != len(newAddresses) {
		return true
	}

	// starts from reverse order
	for i := len(oldAddresses) - 1; i >= 0; i-- {
		var oldAddress = oldAddresses[i]
		var newAddress = newAddresses[i]
		if newAddress.Type != oldAddress.Type || newAddress.Address != oldAddress.Address {
			return true
		}
	}
	return false
}

// diffNodeConditions compares the differences between two Node object's conditions.
// If there is a difference, it returns true.
func diffNodeConditions(oldConditions, newConditions []corev1.NodeCondition) bool {
	if len(oldConditions) != len(newConditions) {
		return true
	}

	// starts from reverse order
	for i := len(oldConditions) - 1; i >= 0; i-- {
		var oldCondition = oldConditions[i]
		var newCondition = newConditions[i]
		if newCondition.Type != oldCondition.Type || newCondition.Status != oldCondition.Status {
			return true
		}
	}
	return false
}
