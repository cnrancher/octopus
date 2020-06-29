package predicate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/rancher/octopus/pkg/util/object"
)

var deviceLinkChangedPredicateLog = ctrl.Log.WithName("predicate").WithName("deviceLinkChanged")

type DeviceLinkChangedPredicate struct {
	predicate.Funcs
}

func (p DeviceLinkChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld == nil || e.MetaNew == nil || e.ObjectNew == nil || e.ObjectOld == nil {
		return false
	}

	// doesn't handle non-DeviceLink object
	if !object.IsDeviceLinkObject(e.ObjectOld) {
		return true
	}

	var dl = object.ToDeviceLinkObject(e.ObjectNew)

	if e.MetaNew.GetGeneration() != e.MetaOld.GetGeneration() {
		if dl.Status.NodeName != dl.Spec.Adaptor.Node {
			deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the node is changed", "object", object.GetNamespacedName(e.MetaOld))
			return true
		}
		if dl.Status.Model != dl.Spec.Model {
			deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the model is changed", "object", object.GetNamespacedName(e.MetaOld))
			return true
		}
		return false
	}

	if dl.GetNodeExistedStatus() == metav1.ConditionUnknown {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the NodeExisted status is checking", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}
	if dl.GetModelExistedStatus() == metav1.ConditionUnknown {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the ModelExisted status is checking", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}

	return false
}
