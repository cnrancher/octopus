package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/limb/index"
	"github.com/rancher/octopus/pkg/limb/predicate"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/fieldpath"
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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

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

	if object.IsDeleted(&link) {
		if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
			return ctrl.Result{}, nil
		}

		// disconnects
		r.SuctionCup.Disconnect(&link)

		// removes finalizer
		link.Finalizers = collection.StringSliceRemove(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "Unable to remove finalizer from DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// rejects if not the requested node
	if link.Status.NodeName != r.NodeName {
		// NB(thxCode) disconnects the link to avoid connection leak when the requested node has been changed
		r.SuctionCup.Disconnect(&link)
		return ctrl.Result{}, nil
	}

	// rejects if the conditions are not met
	if devicelink.GetModelExistedStatus(&link.Status) != metav1.ConditionTrue {
		// NB(thxCode) disconnects the link to avoid connection leak when the model has been changed or removed
		r.SuctionCup.Disconnect(&link)
		return ctrl.Result{}, nil
	}

	// adds finalizer if needed
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
			link.Status.AdaptorName != link.Spec.Adaptor.Name {
			devicelink.ToCheckAdaptorExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if !r.SuctionCup.ExistAdaptor(link.Spec.Adaptor.Name) ||
			link.Status.AdaptorName != link.Spec.Adaptor.Name {
			// NB(thxCode) disconnects the link to avoid connection leak when the requested adaptor has been changed
			r.SuctionCup.Disconnect(&link)
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

		link.Status.AdaptorName = link.Spec.Adaptor.Name
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// validates device created or not
	var device unstructured.Unstructured
	switch devicelink.GetDeviceCreatedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if link.Status.DeviceTemplateGeneration != link.Generation {
			devicelink.ToCheckDeviceCreated(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		var err error

		// makes device from model
		device, err = model.NewInstanceOfTypeMeta(link.Status.Model)
		if err != nil {
			// NB(thxCode) disconnects the link to avoid connection leak when the device instance has not been fetched
			r.SuctionCup.Disconnect(&link)
			devicelink.FailOnDeviceCreated(&link.Status, "unable to make device from typemeta")
			link.Status.DeviceTemplateGeneration = link.Generation
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "FailedCreated", "cannot make device from typemeta: %v", err)
			return ctrl.Result{}, nil
		}

		// fetches device
		if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
			// re-checks if the model exists
			if meta.IsNoMatchError(err) {
				// NB(thxCode) disconnects the link to avoid connection leak when the device instance has not been fetched
				r.SuctionCup.Disconnect(&link)
				devicelink.ToCheckModelExisted(&link.Status)
				if err := r.Status().Update(ctx, &link); err != nil {
					log.Error(err, "Unable to change the status of DeviceLink")
					return ctrl.Result{Requeue: true}, nil
				}
				r.Eventf(&link, "Warning", "NotMatched", "cannot find device model")
				return ctrl.Result{}, nil
			}

			// requeues when occurring any errors except not-found one
			if !apierrs.IsNotFound(err) {
				log.Error(err, "Unable to fetch the device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		// triggers to create the device again if it's not found or deleted accidentally
		if !object.IsActivating(&device) {
			// NB(thxCode) disconnects the link to avoid connection leak when the device instance has not been found
			r.SuctionCup.Disconnect(&link)
			devicelink.ToCheckDeviceCreated(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "Recreating", "cannot find previous device")
			return ctrl.Result{}, nil
		}

		// updates device if needed
		if isDeviceSpecChanged(&link, &device) {
			if err := r.Update(ctx, &device); err != nil {
				// fails if the device is invalid
				if apierrs.IsInvalid(err) {
					// NB(thxCode) disconnects the link to avoid connection leak when the device instance cannot be updated
					r.SuctionCup.Disconnect(&link)
					devicelink.FailOnDeviceCreated(&link.Status, "unable to update device")
					link.Status.DeviceTemplateGeneration = link.Generation
					if err := r.Status().Update(ctx, &link); err != nil {
						log.Error(err, "Unable to change the status of DeviceLink")
						return ctrl.Result{Requeue: true}, nil
					}
					r.Eventf(&link, "Warning", "FailedCreated", "cannot update device from template: %v", err)
					return ctrl.Result{}, nil
				}

				log.Error(err, "Unable to update device")
				return ctrl.Result{Requeue: true}, nil
			}
		}
	default:
		// constructs device from template
		device = constructDevice(&link)

		// creates device
		if err := r.Create(ctx, &device); err != nil {
			// re-checks if the model exists
			if meta.IsNoMatchError(err) {
				devicelink.ToCheckModelExisted(&link.Status)
				if err := r.Status().Update(ctx, &link); err != nil {
					log.Error(err, "Unable to change the status of DeviceLink")
					return ctrl.Result{Requeue: true}, nil
				}
				r.Eventf(&link, "Warning", "NotMatched", "cannot find device model")
				return ctrl.Result{}, nil
			}

			// fails if the device is invalid
			if apierrs.IsInvalid(err) {
				devicelink.FailOnDeviceCreated(&link.Status, "unable to create device from template")
				link.Status.DeviceTemplateGeneration = link.Generation
				if err := r.Status().Update(ctx, &link); err != nil {
					log.Error(err, "Unable to change the status of DeviceLink")
					return ctrl.Result{Requeue: true}, nil
				}
				r.Eventf(&link, "Warning", "FailedCreated", "cannot create device from template: %v", err)
				return ctrl.Result{}, nil
			}

			// requeues when occurring any errors except already-existed one
			if !apierrs.IsAlreadyExists(err) {
				log.Error(err, "Unable to create the device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		devicelink.SuccessOnDeviceCreated(&link.Status)
		link.Status.DeviceTemplateGeneration = link.Generation
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		r.Eventf(&link, "Normal", "Created", "device is created")
		return ctrl.Result{}, nil
	}

	// validates device connected or not
	switch devicelink.GetDeviceConnectedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if link.Status.DeviceTemplateGeneration != link.Generation {
			devicelink.ToCheckDeviceConnected(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "Reconnecting", "triggered by modification")
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		// fetches the device references if needed
		var references, err = r.fetchReferences(ctx, &link)
		if err != nil {
			log.Error(err, "Unable to fetch the reference parameters of DeviceLink")
			r.Eventf(&link, "Warning", "FailedSent", "cannot send data to adaptor as failed to fetch the reference parameters: %v", err)
			return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}

		if err := r.SuctionCup.Send(references, &device, &link); err != nil {
			devicelink.FailOnDeviceConnected(&link.Status, "cannot send data to adaptor")
			link.Status.DeviceTemplateGeneration = link.Generation
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "FailedSent", "cannot send data to adaptor: %v", err)
		}
		return ctrl.Result{}, nil
	default:
		if _, err := r.SuctionCup.Connect(&link); err != nil {
			devicelink.FailOnDeviceConnected(&link.Status, "unable to connect to adaptor")
			link.Status.DeviceTemplateGeneration = link.Generation
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "FailedConnected", "cannot connect to adaptor: %v", err)
			return ctrl.Result{}, nil
		}

		devicelink.WaitForDeviceConnected(&link.Status)
		link.Status.DeviceTemplateGeneration = link.Generation
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		r.Eventf(&link, "Normal", "Connected", "connected to adaptor")
		return ctrl.Result{}, nil
	}
}

func (r *DeviceLinkReconciler) SetupWithManager(ctrlMgr ctrl.Manager, suctionCupMgr suctioncup.Manager) error {
	// registers receiver
	suctionCupMgr.RegisterAdaptorHandler(r)
	suctionCupMgr.RegisterConnectionHandler(r)

	// indexes DeviceLink by `status.adaptorName`
	if err := ctrlMgr.GetFieldIndexer().IndexField(
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByAdaptorField,
		index.DeviceLinkByAdaptorFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(ctrlMgr).
		Named("limb_dl").
		For(&edgev1alpha1.DeviceLink{}).
		WithEventFilter(predicate.DeviceLinkChangedPredicate{NodeName: r.NodeName}).
		Complete(r)
}

// fetchReferences fetches the references of deviceLink.
func (r *DeviceLinkReconciler) fetchReferences(ctx context.Context, deviceLink *edgev1alpha1.DeviceLink) (map[string]map[string][]byte, error) {
	var references = deviceLink.Spec.References
	var namespace = deviceLink.Namespace

	var referencesData map[string]map[string][]byte
	if len(references) != 0 {
		referencesData = make(map[string]map[string][]byte, len(references))

		for _, rp := range references {
			var name = rp.Name

			// fetches secret references
			if rp.Secret != nil {
				var desiredName = rp.Secret.Name
				var desiredItems = rp.Secret.Items

				var secret corev1.Secret
				if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: desiredName}, &secret); err != nil {
					return nil, err
				}

				var items = secret.Data
				if len(desiredItems) != 0 {
					items = make(map[string][]byte, len(desiredItems))
					for _, sk := range desiredItems {
						var sv, exist = secret.Data[sk]
						if !exist {
							return nil, apierrs.NewNotFound(corev1.Resource(corev1.ResourceSecrets.String()), fmt.Sprintf("%s.data(%s)", desiredName, sk))
						}
						items[sk] = sv
					}
				}

				referencesData[name] = items
				continue
			}

			// fetches configMap references
			if rp.ConfigMap != nil {
				var desiredName = rp.ConfigMap.Name
				var desiredItems = rp.ConfigMap.Items

				var configMap corev1.ConfigMap
				if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: desiredName}, &configMap); err != nil {
					return nil, err
				}

				var items map[string][]byte
				if len(desiredItems) != 0 {
					items = make(map[string][]byte, len(desiredItems))
					for _, cmk := range desiredItems {
						var cmv, exist = configMap.Data[cmk]
						if !exist {
							return nil, apierrs.NewNotFound(corev1.Resource(corev1.ResourceConfigMaps.String()), fmt.Sprintf("%s.data(%s)", desiredName, cmk))
						}
						items[cmk] = []byte(cmv)
					}
				} else {
					items = make(map[string][]byte, len(configMap.Data))
					for cmk, cmv := range configMap.Data {
						items[cmk] = []byte(cmv)
					}
				}

				referencesData[name] = items
				continue
			}

			// fetches downward API references
			if rp.DownwardAPI != nil {
				var desiredItems = rp.DownwardAPI.Items

				// the length of items should not be less than 1
				var items = make(map[string][]byte, len(desiredItems))
				for _, dk := range desiredItems {
					var err error
					items[dk.Name], err = fieldpath.ExtractDeviceLinkFieldPathAsBytes(deviceLink, dk.FieldRef.FieldPath)
					if err != nil {
						return nil, apierrs.NewNotFound(edgev1alpha1.GroupResourceDeviceLink, fmt.Sprintf("%s.downwardapi(%s)", deviceLink.Name, dk.FieldRef.FieldPath))
					}
				}

				referencesData[name] = items
			}
		}
	}

	return referencesData, nil
}

// isDeviceSpecChanged returns true if there is any changed from deviceLink's template and applies the changes into device.
func isDeviceSpecChanged(deviceLink *edgev1alpha1.DeviceLink, device *unstructured.Unstructured) bool {
	var deviceTemplate = deviceLink.Spec.Template

	var deviceAnnotationsUpdated = collection.StringMapCopyInto(
		map[string]string{
			"edge.cattle.io/node-name":    deviceLink.Status.NodeName,
			"edge.cattle.io/adaptor-name": deviceLink.Status.AdaptorName,
		},
		collection.StringMapCopy(device.GetAnnotations()))
	var deviceLabelsUpdated = collection.StringMapCopyInto(
		deviceTemplate.Labels,
		collection.StringMapCopy(device.GetLabels()))
	// NB(thxCode) apiserver will take care the format of `template.spec.raw`, so we can consider it as a good JSON format.
	var deviceSpecUpdated map[string]interface{}
	converter.TryUnmarshalJSON(deviceTemplate.Spec.Raw, &deviceSpecUpdated)

	var changed bool
	if collection.DiffStringMap(device.GetAnnotations(), deviceAnnotationsUpdated) {
		changed = true
		device.SetAnnotations(deviceAnnotationsUpdated)
	}
	if collection.DiffStringMap(device.GetLabels(), deviceLabelsUpdated) {
		changed = true
		device.SetLabels(deviceLabelsUpdated)
	}
	if !reflect.DeepEqual(device.Object["spec"], deviceSpecUpdated) {
		changed = true
		device.Object["spec"] = deviceSpecUpdated
	}
	return changed
}

// constructDevice constructs device instance from deviceLink's template.
func constructDevice(deviceLink *edgev1alpha1.DeviceLink) unstructured.Unstructured {
	var deviceModel = deviceLink.Spec.Model
	var deviceTemplate = deviceLink.Spec.Template

	var deviceAnnotations = map[string]string{
		"edge.cattle.io/node-name":    deviceLink.Status.NodeName,
		"edge.cattle.io/adaptor-name": deviceLink.Status.AdaptorName,
	}
	var deviceLabels = collection.StringMapCopy(deviceTemplate.Labels)
	// NB(thxCode) apiserver will take care the format of `template.spec.raw`, so we can consider it as a good JSON format.
	var deviceSpec map[string]interface{}
	converter.TryUnmarshalJSON(deviceTemplate.Spec.Raw, &deviceSpec)

	var device = unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       deviceModel.Kind,
			"apiVersion": deviceModel.APIVersion,
			"metadata": map[string]interface{}{
				"name":        deviceLink.Name,
				"namespace":   deviceLink.Namespace,
				"labels":      deviceLabels,
				"annotations": deviceAnnotations,
			},
			"spec": deviceSpec,
		},
	}
	device.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(deviceLink, deviceLink.GroupVersionKind()),
	})
	return device
}
