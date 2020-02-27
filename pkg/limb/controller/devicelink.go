package controller

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	jsoniter "github.com/json-iterator/go"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor"
	"github.com/rancher/octopus/pkg/adaptor/registration"
	"github.com/rancher/octopus/pkg/limb/index"
	"github.com/rancher/octopus/pkg/limb/model"
	"github.com/rancher/octopus/pkg/status"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/log/handler"
	"github.com/rancher/octopus/pkg/util/object"
)

const (
	ReconcilingDeviceLink = "edge.cattle.io/octopus-limb"
)

// DeviceLinkReconciler reconciles a DeviceLink object
type DeviceLinkReconciler struct {
	client.Client
	record.EventRecorder

	Scheme *k8sruntime.Scheme
	Log    logr.Logger

	Adaptors adaptor.Pool
	NodeName string
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *DeviceLinkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = context.Background()
	var log = r.Log.WithValues("devicelink", req.NamespacedName)

	defer func() {
		log.V(0).Info("reconcile out")
	}()
	log.V(0).Info("reconcile in")

	// fetches link
	var link edgev1alpha1.DeviceLink
	if err := r.Get(ctx, req.NamespacedName, &link); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}

	// validates application
	if link.Spec.Adaptor.Node != r.NodeName {
		return ctrl.Result{}, nil
	}

	// validates mode existing or not
	if status.GetModelExistedStatus(&link.Status) != metav1.ConditionTrue {
		log.V(0).Info("model isn't existed", "model", link.Spec.Model)
		return ctrl.Result{}, nil
	}

	if object.IsDeleted(&link) {
		if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
			return ctrl.Result{}, nil
		}

		// disconnect
		if err := r.Adaptors.DeleteConnection(&link); err != nil {
			log.Error(err, "unable to delete the connection between device and DeviceLink")
		}

		// remove finalizer
		link.Finalizers = collection.StringSliceRemove(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "unable to remove finalizer from DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// add finalizer if needed
	if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
		link.Finalizers = append(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "unable to add finalizer to DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// validates adaptor
	switch status.GetAdaptorExistedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if r.Adaptors.Exist(link.Spec.Adaptor.Name) || link.Spec.Adaptor.Name != link.Status.Adaptor.Name ||
			bytes.Compare(link.Spec.Adaptor.Parameters.Raw, link.Status.Adaptor.Parameters.Raw) != 0 {
			status.ToCheckAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if !r.Adaptors.Exist(link.Spec.Adaptor.Name) ||
			link.Spec.Adaptor.Name != link.Status.Adaptor.Name ||
			bytes.Compare(link.Spec.Adaptor.Parameters.Raw, link.Status.Adaptor.Parameters.Raw) != 0 {
			status.ToCheckAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
	default:
		if r.Adaptors.Exist(link.Spec.Adaptor.Name) {
			status.SuccessOnAdaptorExisted(&link.Status)
		} else {
			status.FailOnAdaptorExisted(&link.Status, "the adaptor isn't existed")
		}

		link.Status.Adaptor.Name = link.Spec.Adaptor.Name
		link.Status.Adaptor.Parameters = link.Spec.Adaptor.Parameters
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// creates device instance
	//
	//┌ ─ ─ ─ ─ ─ ─ ─ ┐     F                 S
	//      False      ◀────── Create Device ───────────────┐
	//└ ─ ─ ─ ─ ─ ─ ─ ┘              ▲                      │
	//        │                      │                      │
	//        │                      │                      │
	//        │                      │ N                    │
	//        │                      │        Y             │
	//        ├──────────────▶ Device Exist? ───────────────┤
	//        │                                             ▼
	//┌ ─ ─ ─ ─ ─ ─ ─ ┐                             ┌ ─ ─ ─ ─ ─ ─ ─ ┐
	//     Unknown                                        True
	//└ ─ ─ ─ ─ ─ ─ ─ ┘                             └ ─ ─ ─ ─ ─ ─ ─ ┘
	//        ▲              N                              │
	//        └─────────────── Device Exist? ◀──────────────┘
	//                               │
	//                               │ Y
	//                               │
	//                               │
	//                               ▼
	var device = model.NewInstanceOfTypeMeta(link.Spec.Model)
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch the device of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}
	if status.GetDeviceConnectedStatus(&link.Status) != metav1.ConditionTrue {
		if !object.IsActivating(&device) {
			// create device instance
			if device, err := constructDeviceFromDeviceLink(&link, r.Scheme); err != nil {
				log.Error(err, "unable to construct the device of DeviceLink")
				status.FailOnDeviceConnected(&link.Status, fmt.Sprintf("failed to construct device: %v", err))
			} else if err := r.Create(ctx, &device); err != nil {
				log.Error(err, "unable to create the device of DeviceLink")
				status.FailOnDeviceConnected(&link.Status, fmt.Sprintf("failed to create device: %v", err))
			} else {
				return ctrl.Result{Requeue: true}, nil
			}
		} else if err := r.Adaptors.CreateConnection(&link); err != nil {
			log.Error(err, "unable to create the connection between device and DeviceLink")
			status.FailOnDeviceConnected(&link.Status, fmt.Sprintf("failed to connect device: %v", err))
		} else {
			status.SuccessOnDeviceConnected(&link.Status)
		}

		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	} else {
		if !object.IsActivating(&device) {
			status.ToCheckDeviceConnected(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
	}

	// sync device
	var deviceUpdate = device.DeepCopy()
	var deviceTemplate, err = k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&link.Spec.Template)
	if err != nil {
		log.Error(err, "unable to convert from DeviceLink template")
		r.Eventf(&link, "Warning", "FailedReconcile", "unable to convert device from DeviceLink template: %v", err)
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}
	var deviceLabels, _, _ = unstructured.NestedStringMap(deviceTemplate, "metadata", "labels")
	_ = unstructured.SetNestedStringMap(deviceUpdate.Object, deviceLabels, "metadata", "labels")
	var deviceAnnotations, _, _ = unstructured.NestedStringMap(deviceTemplate, "metadata", "annotations")
	_ = unstructured.SetNestedStringMap(deviceUpdate.Object, fillDeviceAnnotations(&link, deviceAnnotations), "metadata", "annotations")
	var deviceSpec, _, _ = unstructured.NestedMap(deviceTemplate, "spec")
	_ = unstructured.SetNestedMap(deviceUpdate.Object, deviceSpec, "spec")
	if !reflect.DeepEqual(&device, deviceUpdate) {
		if err := r.Adaptors.SendDataToConnection(&link, deviceUpdate); err != nil {
			log.Error(err, "unable to send data to the device of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}

		if err := r.Update(ctx, deviceUpdate); err != nil {
			log.Error(err, "unable to update the device of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *DeviceLinkReconciler) ReceiveAdaptorRegistration(adaptorName string, notice registration.Event, msg string) {
	var ctx = context.Background()
	var log = r.Log.WithValues("adaptor", adaptorName)

	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByAdaptorField: adaptorName}); err != nil {
		log.Error(err, "unable to list related DeviceLink of adaptor")
		return
	}

	switch notice {
	case registration.EventStarted:
		// move link AdaptorExisted condition from `False` to `True`
		for _, link := range links.Items {
			if status.GetAdaptorExistedStatus(&link.Status) != metav1.ConditionFalse {
				continue
			}
			status.SuccessOnAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
			}
		}
	case registration.EventStopped:
		// move link AdaptorExisted condition to `False`
		for _, link := range links.Items {
			if status.GetAdaptorExistedStatus(&link.Status) == metav1.ConditionFalse {
				continue
			}
			status.FailOnAdaptorExisted(&link.Status, "the adaptor is stopped")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
			}
		}
	}
}

func (r *DeviceLinkReconciler) ReceiveDeviceChanges(adaptorName string, deviceObserved *unstructured.Unstructured, err error) {
	var ctx = context.Background()
	var log = r.Log.WithValues("adaptor", adaptorName)

	if err != nil {
		log.Error(err, "receiving error from connection")
		return
	}

	var deviceKey = object.GetNamespacedName(deviceObserved)
	var device = model.NewInstanceOfType(deviceObserved)
	if err := r.Get(ctx, deviceKey, &device); err != nil {
		log.Error(err, "unable to fetch the device", "device", deviceKey)
		return
	}

	var deviceStatus, _, _ = unstructured.NestedMap(deviceObserved.Object, "status")
	_ = unstructured.SetNestedMap(device.Object, deviceStatus, "status")
	if err := r.Status().Update(ctx, &device); err != nil {
		log.Error(err, "unable to update the device", "device", deviceKey)
	}
}

func (r *DeviceLinkReconciler) SetupWithManager(name string, ctrlMgr ctrl.Manager, adaptorMgr adaptor.Manager) error {
	// registers receiver
	adaptorMgr.RegisterConnectionReceiver(r)
	adaptorMgr.RegisterRegistrationReceiver(r)

	// indexes DeviceLink by `spec.adaptor.name`
	if err := ctrlMgr.GetFieldIndexer().IndexField(
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByAdaptorField,
		index.DeviceLinkByAdaptorFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(ctrlMgr).
		Named(name + ".DeviceLink").
		For(&edgev1alpha1.DeviceLink{}). // TODO only watch links of the same node
		Complete(r)
}

func constructDeviceFromDeviceLink(link *edgev1alpha1.DeviceLink, scheme *k8sruntime.Scheme) (unstructured.Unstructured, error) {
	var deviceModel = link.Spec.Model
	var deviceTemplate = link.Spec.Template
	var deviceLabels = collection.StringMapCopy(deviceTemplate.Labels)
	var deviceAnnotations = fillDeviceAnnotations(link, collection.StringMapCopy(deviceTemplate.Annotations))
	var deviceSpec map[string]interface{}
	if err := jsoniter.Unmarshal(deviceTemplate.Spec.Raw, &deviceSpec); err != nil {
		return unstructured.Unstructured{}, err
	}

	var device = unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       deviceModel.Kind,
			"apiVersion": deviceModel.APIVersion,
			"metadata": map[string]interface{}{
				"name":        link.Name,
				"namespace":   link.Namespace,
				"labels":      deviceLabels,
				"annotations": deviceAnnotations,
			},
			"spec": deviceSpec,
		},
	}
	if err := ctrl.SetControllerReference(link, &device, scheme); err != nil {
		return unstructured.Unstructured{}, err
	}
	return device, nil
}

func fillDeviceAnnotations(link *edgev1alpha1.DeviceLink, deviceAnnotations map[string]string) map[string]string {
	if deviceAnnotations == nil {
		deviceAnnotations = make(map[string]string)
	}
	var deviceAdaptor = link.Spec.Adaptor
	deviceAnnotations["edge.cattle.io/adaptor-node"] = deviceAdaptor.Node
	deviceAnnotations["edge.cattle.io/adaptor-name"] = deviceAdaptor.Name
	deviceAnnotations["edge.cattle.io/adaptor-parameters"] = string(deviceAdaptor.Parameters.Raw)
	return deviceAnnotations
}
