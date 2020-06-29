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

	NodeName string
}

func (p DeviceLinkChangedPredicate) Create(e event.CreateEvent) bool {
	if e.Meta == nil || e.Object == nil {
		return false
	}

	// doesn't handle non-DeviceLink object
	if !object.IsDeviceLinkObject(e.Object) {
		return true
	}

	var dl = object.ToDeviceLinkObject(e.Object)

	// NB(thxCode) when creating a fresh new DeviceLink, brain will confirm the Node and fill the result to `status.nodeName`,
	// limb won't accept any create events, but the first list of `list-watch` is parsed as create event,
	// so we need to predicate the related items.
	if dl.Status.NodeName == p.NodeName {
		deviceLinkChangedPredicateLog.V(5).Info("Accept CreateEvent as requested the same node", "object", object.GetNamespacedName(e.Meta))
		return true
	}

	return false
}

func (p DeviceLinkChangedPredicate) Delete(e event.DeleteEvent) bool {
	if e.Meta == nil || e.Object == nil {
		return false
	}

	// doesn't handle non-DeviceLink object
	if !object.IsDeviceLinkObject(e.Object) {
		return true
	}

	// NB(thxCode) there is a finalizer to handler the DeviceLink deletion event,
	// so with the finalizer, the deletion event can be changed to an update event.
	return false
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
	// rejects if not the target node
	if dl.Spec.Adaptor.Node != p.NodeName && dl.Status.NodeName != p.NodeName {
		return false
	}

	if e.MetaNew.GetGeneration() != e.MetaOld.GetGeneration() {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the object's generation is changed", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}

	if dl.GetNodeExistedStatus() == metav1.ConditionFalse {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the NodeExisted status is failed", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}
	if dl.GetModelExistedStatus() == metav1.ConditionFalse {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the ModelExisted status is failed", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}
	if dl.GetAdaptorExistedStatus() == metav1.ConditionUnknown {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the AdaptorExisted status is checking", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}
	if dl.GetDeviceCreatedStatus() == metav1.ConditionUnknown {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the DeviceCreated status is checking", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}
	if dl.GetDeviceConnectedStatus() == metav1.ConditionUnknown {
		deviceLinkChangedPredicateLog.V(5).Info("Accept UpdateEvent as the DeviceConnected status is checking", "object", object.GetNamespacedName(e.MetaOld))
		return true
	}

	return false
}
