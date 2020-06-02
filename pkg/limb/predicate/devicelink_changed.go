package predicate

import (
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

	// handles if the requested node
	if p.NodeName == dl.Spec.Adaptor.Node {
		deviceLinkChangedPredicateLog.V(0).Info("Accept CreateEvent as requested the same node", "key", object.GetNamespacedName(e.Meta))
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

	var dlOld = object.ToDeviceLinkObject(e.ObjectOld)
	var dlNew = object.ToDeviceLinkObject(e.ObjectNew)

	// handles if the object is requesting the same node
	if p.NodeName == dlNew.Status.NodeName {
		deviceLinkChangedPredicateLog.V(0).Info("Accept UpdateEvent as the object is requesting the same node", "key", object.GetNamespacedName(e.MetaNew))
		return true
	}

	// handles if the object has requested the same node previously
	if p.NodeName == dlOld.Status.NodeName {
		// NB(thxCode) help the reconciling logic to close the previous connection
		deviceLinkChangedPredicateLog.V(0).Info("Accept UpdateEvent as the object has requested the same node previously", "key", object.GetNamespacedName(e.MetaOld))
		return true
	}

	return false
}

func (p DeviceLinkChangedPredicate) Generic(e event.GenericEvent) bool {
	if e.Meta == nil || e.Object == nil {
		return false
	}

	// doesn't handle non-DeviceLink object
	if !object.IsDeviceLinkObject(e.Object) {
		return true
	}

	var dl = object.ToDeviceLinkObject(e.Object)

	// handles if the requested node
	if p.NodeName == dl.Spec.Adaptor.Node {
		deviceLinkChangedPredicateLog.V(0).Info("Accept GenericEvent as requested the same node", "key", object.GetNamespacedName(e.Meta))
		return true
	}

	return false
}
