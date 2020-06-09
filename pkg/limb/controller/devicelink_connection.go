package controller

import (
	"context"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=list
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *DeviceLinkReconciler) ReceiveConnectionStatus(req suctioncup.RequestConnectionStatus) (suctioncup.Response, error) {
	var ctx = context.Background()
	var log = r.Log.WithName("connectionReceiving").WithValues("adaptor", req.AdaptorName, "deviceLink", req.Name)

	// validates link
	var link edgev1alpha1.DeviceLink
	if err := r.Get(ctx, req.Name, &link); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		// NB(thxCode) the received data of the stale link can be discarded
		return suctioncup.Response{}, nil
	}
	switch devicelink.GetDeviceConnectedStatus(&link.Status) {
	case metav1.ConditionFalse, metav1.ConditionTrue:
	default:
		// NB(thxCode) if the connected status of device is unknown,
		// the received data can be discarded
		return suctioncup.Response{}, nil
	}

	// validates device
	target, err := model.NewInstanceOfTypeMeta(link.Status.Model)
	if err != nil {
		log.Error(err, "Unable to get device of DeviceLink")
		return suctioncup.Response{}, nil
	}
	if err := r.Get(ctx, req.Name, &target); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to get the device of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
	}
	if !object.IsActivating(&target) {
		devicelink.ToCheckDeviceCreated(&link.Status)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "Recreating", "cannot find previous device")
		// NB(thxCode) if the device is not activated,
		// the current connection must be recreated,
		// so the received data can be discarded
		return suctioncup.Response{}, nil
	}

	if req.Closed {
		// NB(thxCode) if the connection is closed passively, we need to reconnect again.
		// However, we need a way to stop the passive closed from unregistering adaptor. TODO
		devicelink.ToCheckDeviceConnected(&link.Status)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "Reconnecting", "disconnected by adaptor")
		return suctioncup.Response{}, nil
	}

	if req.Error != nil {
		// NB(thxCode) if the connection returns an error, it may be something uncontrollable happened,
		// e.g. passed a wrong parameter or failed to connect the physical device, so we cannot reconnect directly.
		// the connection can be resumed via user modified.
		devicelink.FailOnDeviceConnected(&link.Status, "received error from adaptor")
		link.Status.DeviceTemplateGeneration = link.Generation
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "FailedReceived", "received error from adaptor: %v", req.Error)
		return suctioncup.Response{}, nil
	}

	// updates device status
	var updatedStatus interface{}
	if err := func() error {
		var updated = &unstructured.Unstructured{Object: make(map[string]interface{})}
		if err := updated.UnmarshalJSON(req.Data); err != nil {
			return err
		}
		updatedStatus = updated.Object["status"]
		return nil
	}(); err != nil {
		// NB(thxCode) if failed to process data, we just record an event for this.
		r.Eventf(&link, "Warning", "FailReceived", "received invalid data from adaptor: %v", err)
		return suctioncup.Response{}, nil
	}

	target.Object["status"] = updatedStatus
	if err := r.Status().Update(ctx, &target); err != nil {
		log.Error(err, "Unable to update the device of DeviceLink")
		return suctioncup.Response{Requeue: true}, nil
	}
	return suctioncup.Response{}, nil
}
