package controller

import (
	"context"
	"fmt"

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
	var log = r.Log.WithName("ReceiveConnectionStatus").WithValues("adaptor", req.AdaptorName, "devicelink", req.Name)

	// validates link
	var link edgev1alpha1.DeviceLink
	if err := r.Get(ctx, req.Name, &link); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch DeviceLink")
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
	var target = model.NewInstanceOfTypeMeta(link.Status.Model)
	if err := r.Get(ctx, req.Name, &target); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to the device of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
	}
	if !object.IsActivating(&target) {
		devicelink.ToCheckDeviceCreated(&link.Status)
		r.Eventf(&link, "Warning", "Recreating", "previous device instance is gone")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		// NB(thxCode) if the device is not activated,
		// the current connection must be recreated,
		// so the received data can be discarded
		return suctioncup.Response{}, nil
	}

	if req.Closed {
		devicelink.ToCheckDeviceConnected(&link.Status)
		r.Eventf(&link, "Warning", "Reconnecting", "previous connection is stopped")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		return suctioncup.Response{}, nil
	}

	if req.Error != nil {
		devicelink.FailOnDeviceConnected(&link.Status, fmt.Sprintf("received error: %v", req.Error.Error()))
		r.Eventf(&link, "Warning", "FailReceivedData", "received error")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
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
		devicelink.FailOnDeviceConnected(&link.Status, fmt.Sprintf("received data is invalid: %v", err.Error()))
		r.Eventf(&link, "Warning", "FailReceivedData", "received data is invalid")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		return suctioncup.Response{}, nil
	}

	target.Object["status"] = updatedStatus
	if err := r.Status().Update(ctx, &target); err != nil {
		log.Error(err, "unable to update the device of DeviceLink")
		return suctioncup.Response{Requeue: true}, nil
	}
	return suctioncup.Response{}, nil
}
