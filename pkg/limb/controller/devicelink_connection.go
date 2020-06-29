package controller

import (
	"context"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/log/handler"
	modelutil "github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=list
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *DeviceLinkReconciler) ReceiveConnectionStatus(req suctioncup.RequestConnectionStatus) (suctioncup.Response, error) {
	var ctx = context.Background()
	var log = r.Log.WithName("connectionReceiving").WithValues("adaptor", req.AdaptorName, "deviceLink", req.Name)

	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	// validates link
	var link edgev1alpha1.DeviceLink
	if err := r.Get(ctx, req.Name, &link); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		// NB(thxCode) we just discard the received data in this case.
		return suctioncup.Response{}, nil
	}

	if req.Closed {
		// NB(thxCode) we need to reconnect again if the connection is closed passively.
		// TODO However, we need a way to stop the passive closed from unregistering adaptor.
		link.ToCheckDeviceConnected()
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "Reconnecting", "closed by adaptor")
		return suctioncup.Response{}, nil
	}

	if req.Error != nil {
		// NB(thxCode) we cannot reconnect directly if the connection returns an error,
		// it may be something uncontrollable happened, e.g. passed a wrong parameter or failed to connect the physical device.
		// it can be recovered by user manually.
		r.SuctionCup.Disconnect(&link)
		link.FailOnDeviceConnected("received error from adaptor")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "Disconnected", "received error from adaptor: %v", req.Error)
		return suctioncup.Response{}, nil
	}

	// moves next if success on DeviceConnected
	if link.GetDeviceConnectedStatus() != metav1.ConditionTrue {
		return suctioncup.Response{}, nil
	}

	// parses device status
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

	// validates device
	var device, err = modelutil.NewInstanceOfTypeMeta(link.Status.Model)
	if err != nil {
		// NB(thxCode) we don't need to deal with this case as it can be traced by the main logic of limb.
		return suctioncup.Response{}, nil
	}
	if err := r.Get(ctx, req.Name, &device); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to get the device of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
	}
	if !object.IsActivating(&device) {
		// NB(thxCode) we should trigger to recreate the device if it is deleted.
		link.ToCheckDeviceCreated()
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "Recreating", "previous device is inactivated")
		return suctioncup.Response{}, nil
	}
	device.Object["status"] = updatedStatus
	if err := r.Status().Update(ctx, &device); err != nil {
		log.Error(err, "Unable to update the device of DeviceLink")
		return suctioncup.Response{Requeue: true}, nil
	}

	return suctioncup.Response{}, nil
}
