package controller

import (
	"bytes"
	"context"
	"reflect"

	"github.com/go-logr/logr"
	jsoniter "github.com/json-iterator/go"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/limb/index"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/model"
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

	SuctionCup suctioncup.Neurons
	NodeName   string
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *DeviceLinkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = context.Background()
	var log = r.Log.WithValues("deviceLink", req.NamespacedName)

	// fetches link
	var link edgev1alpha1.DeviceLink
	if err := r.Get(ctx, req.NamespacedName, &link); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}

	// validates application
	if link.Status.Adaptor.Node != r.NodeName {
		return ctrl.Result{}, nil
	}

	// validates mode existing or not
	if devicelink.GetModelExistedStatus(&link.Status) != metav1.ConditionTrue {
		return ctrl.Result{}, nil
	}

	if object.IsDeleted(&link) {
		if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
			return ctrl.Result{}, nil
		}

		// disconnect
		r.SuctionCup.Disconnect(&link)

		// remove finalizer
		link.Finalizers = collection.StringSliceRemove(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "Unable to remove finalizer from DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// add finalizer if needed
	if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
		link.Finalizers = append(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "Unable to add finalizer to DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// validates adaptor existing or not
	switch devicelink.GetAdaptorExistedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if r.SuctionCup.ExistAdaptor(link.Spec.Adaptor.Name) ||
			link.Spec.Adaptor.Name != link.Status.Adaptor.Name ||
			compareAdaptorParameters(link.Spec.Adaptor, link.Status.Adaptor) {
			devicelink.ToCheckAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if !r.SuctionCup.ExistAdaptor(link.Spec.Adaptor.Name) ||
			link.Spec.Adaptor.Name != link.Status.Adaptor.Name ||
			compareAdaptorParameters(link.Spec.Adaptor, link.Status.Adaptor) {
			devicelink.ToCheckAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
	default:
		if r.SuctionCup.ExistAdaptor(link.Spec.Adaptor.Name) {
			devicelink.SuccessOnAdaptorExisted(&link.Status)
		} else {
			devicelink.FailOnAdaptorExisted(&link.Status, "the adaptor isn't existed")
		}

		link.Status.Adaptor.Name = link.Spec.Adaptor.Name
		link.Status.Adaptor.Parameters = link.Spec.Adaptor.Parameters
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// validate device created or not
	var device unstructured.Unstructured
	switch devicelink.GetDeviceCreatedStatus(&link.Status) {
	case metav1.ConditionFalse:
		// TODO use the admission webhook to transfer this
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		var err error
		device, err = model.NewInstanceOfTypeMeta(link.Status.Model)
		if err != nil {
			devicelink.FailOnDeviceCreated(&link.Status, "unable to update device from template")
			r.Eventf(&link, "Warning", "FailedCreated", "cannot update device from template: %v", err)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
		if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
			if !apierrs.IsNotFound(err) && !meta.IsNoMatchError(err) {
				log.Error(err, "Unable to fetch the device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		if !object.IsActivating(&device) {
			devicelink.ToCheckDeviceCreated(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}

		// update device
		updated, err := updateDevice(&link, &device)
		if err != nil {
			devicelink.FailOnDeviceCreated(&link.Status, "unable to update device from template")
			r.Eventf(&link, "Warning", "FailedCreated", "cannot update device from template: %v", err)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
		if updated {
			if err := r.Update(ctx, &device); err != nil {
				log.Error(err, "Failed to update device")
				return ctrl.Result{Requeue: true}, nil
			}
		}
	default:
		// create device
		if device, err := constructDevice(&link, r.Scheme); err != nil {
			devicelink.FailOnDeviceCreated(&link.Status, "unable to construct device from template")
			r.Eventf(&link, "Warning", "FailedCreated", "cannot create device from template: %v", err)
		} else {
			var err = r.Create(ctx, &device)
			if err != nil {
				if !apierrs.IsAlreadyExists(err) {
					log.Error(err, "Unable to create the device of DeviceLink")
					return ctrl.Result{Requeue: true}, nil
				}
			}
			if meta.IsNoMatchError(err) {
				devicelink.FailOnDeviceCreated(&link.Status, "unable to construct device from template")
				r.Eventf(&link, "Warning", "FailedCreated", "cannot create device from template: the model isn't existed")
			} else {
				devicelink.SuccessOnDeviceCreated(&link.Status)
				r.Eventf(&link, "Normal", "Created", "device instance is created")
			}
		}

		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// validate device connected or not
	switch devicelink.GetDeviceConnectedStatus(&link.Status) {
	case metav1.ConditionFalse:
		// NB(thxCode) could not send any data to unhealthy connection,
		// this status changes maybe can drive by suction cup.
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if err := r.SuctionCup.Send(&device, &link); err != nil {
			devicelink.FailOnDeviceConnected(&link.Status, "cannot send data to adaptor")
			r.Eventf(&link, "Warning", "FailedSent", "cannot send data to adaptor: %v", err)

			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	default:
		if err := r.SuctionCup.Connect(&link); err != nil {
			devicelink.FailOnDeviceConnected(&link.Status, "unable to connect to adaptor")
			r.Eventf(&link, "Warning", "FailedConnected", "cannot connect to adaptor: %v", err)
		} else {
			devicelink.SuccessOnDeviceConnected(&link.Status)
			r.Eventf(&link, "Normal", "Connected", "connected to adaptor")
		}

		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}
}

func (r *DeviceLinkReconciler) SetupWithManager(ctrlMgr ctrl.Manager, suctionCupMgr suctioncup.Manager) error {
	// registers receiver
	suctionCupMgr.RegisterAdaptorHandler(r)
	suctionCupMgr.RegisterConnectionHandler(r)

	// indexes DeviceLink by `spec.adaptor.name`
	if err := ctrlMgr.GetFieldIndexer().IndexField(
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByAdaptorField,
		index.DeviceLinkByAdaptorFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(ctrlMgr).
		Named("DeviceLink").
		For(&edgev1alpha1.DeviceLink{}).
		Complete(r)
}

func updateDevice(from *edgev1alpha1.DeviceLink, target *unstructured.Unstructured) (updated bool, err error) {
	var original = target.DeepCopy()

	var updatedAnnotations = markDevice(from, target.GetAnnotations())
	var updatedLabels map[string]string
	var updatedSpec map[string]interface{}
	if err := func() error {
		var template = from.Spec.Template
		if err := jsoniter.Unmarshal(template.Spec.Raw, &updatedSpec); err != nil {
			return err
		}
		updatedLabels = collection.StringMapCopyInto(template.Labels, target.GetLabels())
		return nil
	}(); err != nil {
		return false, err
	}

	target.SetLabels(updatedLabels)
	target.SetAnnotations(updatedAnnotations)
	target.Object["spec"] = updatedSpec
	// another way to update spec:
	// _ = unstructured.SetNestedMap(target.Object, updatedSpec, "spec")
	return !reflect.DeepEqual(target, original), nil
}

func constructDevice(from *edgev1alpha1.DeviceLink, scheme *k8sruntime.Scheme) (unstructured.Unstructured, error) {
	var deviceModel = from.Status.Model
	var deviceName = from.Name
	var deviceNamespace = from.Namespace
	var deviceAnnotations = markDevice(from, nil)
	var deviceLabels map[string]string
	var deviceSpec map[string]interface{}
	if err := func() error {
		var template = from.Spec.Template
		if err := jsoniter.Unmarshal(template.Spec.Raw, &deviceSpec); err != nil {
			return err
		}
		deviceLabels = collection.StringMapCopy(template.Labels)
		return nil
	}(); err != nil {
		return unstructured.Unstructured{}, err
	}

	var device = unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       deviceModel.Kind,
			"apiVersion": deviceModel.APIVersion,
			"metadata": map[string]interface{}{
				"name":        deviceName,
				"namespace":   deviceNamespace,
				"labels":      deviceLabels,
				"annotations": deviceAnnotations,
			},
			"spec": deviceSpec,
		},
	}
	if err := ctrl.SetControllerReference(from, &device, scheme); err != nil {
		return unstructured.Unstructured{}, err
	}
	return device, nil
}

func markDevice(link *edgev1alpha1.DeviceLink, deviceAnnotations map[string]string) map[string]string {
	if deviceAnnotations == nil {
		deviceAnnotations = make(map[string]string)
	}
	var deviceAdaptor = link.Spec.Adaptor
	deviceAnnotations["edge.cattle.io/adaptor-node"] = deviceAdaptor.Node
	deviceAnnotations["edge.cattle.io/adaptor-name"] = deviceAdaptor.Name
	if deviceAdaptor.Parameters != nil {
		deviceAnnotations["edge.cattle.io/adaptor-parameters"] = string(deviceAdaptor.Parameters.Raw)
	}
	return deviceAnnotations
}

func compareAdaptorParameters(adaptor, statusAdaptor edgev1alpha1.DeviceAdaptor) bool {
	if adaptor.Parameters == nil || statusAdaptor.Parameters == nil {
		return false
	}
	return bytes.Compare(adaptor.Parameters.Raw, statusAdaptor.Parameters.Raw) != 0
}
