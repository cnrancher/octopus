package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/limb/index"
	"github.com/rancher/octopus/pkg/limb/predicate"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/fieldpath"
	modelutil "github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

const (
	ReconcilingDeviceLink = "edge.cattle.io/octopus-limb"
)

// DeviceLinkReconciler reconciles a DeviceLink object
type DeviceLinkReconciler struct {
	client.Client
	record.EventRecorder

	Ctx context.Context
	Log logr.Logger

	SuctionCup suctioncup.Neurons
	NodeName   string
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func (r *DeviceLinkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = r.Ctx
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

	// adds finalizer if needed
	if !collection.StringSliceContain(link.Finalizers, ReconcilingDeviceLink) {
		link.Finalizers = append(link.Finalizers, ReconcilingDeviceLink)
		if err := r.Update(ctx, &link); err != nil {
			log.Error(err, "Unable to add finalizer to DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// NB(thxCode) we might see this as the `spec.adaptor.node` has been changed,
	// so we need to disconnect the previous connection and
	// wait for brain to confirm the next step.
	if link.Status.NodeName != link.Spec.Adaptor.Node {
		r.SuctionCup.Disconnect(&link)
		return ctrl.Result{}, nil
	}

	// NB(thxCode) we might see this as the `spec.model` has been changed,
	// so we need to disconnect the previous connection and
	// wait for brain to confirm the next step.
	if link.Status.Model != link.Spec.Model {
		r.SuctionCup.Disconnect(&link)
		return ctrl.Result{}, nil
	}

	// NB(thxCode) we might see this as the `spec.adaptor.name` has been changed,
	// so we need to disconnect the previous connection.
	if link.Status.AdaptorName != link.Spec.Adaptor.Name {
		r.SuctionCup.Disconnect(&link)
	}

	// validates adaptor
	var isAdaptorExisted = r.SuctionCup.ExistAdaptor(link.Spec.Adaptor.Name)
	if !isAdaptorExisted {
		link.FailOnAdaptorExisted("the adaptor isn't existed")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}
	link.SucceedOnAdaptorExisted()

	// validates device
	var device, deviceNewErr = modelutil.NewInstanceOfTypeMeta(link.Status.Model)
	if deviceNewErr != nil {
		log.Error(deviceNewErr, "Unable to make device from model")
		link.FailOnDeviceCreated("unable to make device from model")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		// requeues when occurring any errors except not-found one
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch the device of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}
	if !object.IsActivating(&device) {
		// creates device
		var deviceNew = constructDeviceFromTemplate(&link)
		if err := r.Create(ctx, &deviceNew); err != nil {
			if !apierrs.IsInvalid(err) {
				log.Error(err, "Unable to create the device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}

			// NB(thxCode) if the device creation is invalid, we don't need to retry.
			log.Error(err, "Unable to create device from template")
			link.FailOnDeviceCreated("unable to create device from template")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}

		device = deviceNew
	}
	link.SucceedOnDeviceCreated()

	// fetches the references
	var references, err = r.fetchReferences(&link)
	if err != nil {
		link.FailOnDeviceConnected("unable to fetch the reference parameters")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "FailedFetched", "cannot fetch the reference parameters: %v, retry in 10 seconds", err)
		// NB(thxCode) since we don't list-watch the ConfigMap/Secret resources, we have to give a retry mechanism to obtain these resources.
		return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	// updates device if need
	var updateDevice = isDeviceSpecChanged(&link, &device)
	if updateDevice {
		if err := r.Update(ctx, &device); err != nil {
			if !apierrs.IsInvalid(err) {
				log.Error(err, "Unable to update the device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			link.FailOnDeviceConnected("unable to update the device from template")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			r.Eventf(&link, "Warning", "FailedUpdated", "cannot update the device from template: %v", err)
			return ctrl.Result{}, nil
		}
	}

	// connects to device
	if err := r.SuctionCup.Connect(references, &device, &link); err != nil {
		link.FailOnDeviceConnected("unable to connect to device")
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		r.Eventf(&link, "Warning", "FailedConnected", "cannot connect to device: %v", err)
		return ctrl.Result{}, nil
	}
	link.SucceedOnDeviceConnected()

	if err := r.Status().Update(ctx, &link); err != nil {
		log.Error(err, "Unable to change the status of DeviceLink")
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *DeviceLinkReconciler) SetupWithManager(ctrlMgr ctrl.Manager, suctionCupMgr suctioncup.Manager) error {
	// registers receiver
	suctionCupMgr.RegisterAdaptorHandler(r)
	suctionCupMgr.RegisterConnectionHandler(r)

	if err := ctrlMgr.GetFieldIndexer().IndexField(
		r.Ctx,
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByAdaptorField,
		index.DeviceLinkByAdaptorFuncFactory(r.NodeName),
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
func (r *DeviceLinkReconciler) fetchReferences(deviceLink *edgev1alpha1.DeviceLink) (map[string]map[string][]byte, error) {
	var ctx = r.Ctx
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
	var deviceSpecUpdated = make(map[string]interface{}, 0)
	if deviceTemplate.Spec != nil {
		// NB(thxCode) apiserver will take care the format of `template.spec.raw`,
		// so we can consider it as a good JSON format.
		converter.TryUnmarshalJSON(deviceTemplate.Spec.Raw, &deviceSpecUpdated)
	}

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

// constructDeviceFromTemplate constructs device instance from deviceLink's template.
func constructDeviceFromTemplate(deviceLink *edgev1alpha1.DeviceLink) unstructured.Unstructured {
	var deviceModel = deviceLink.Spec.Model
	var deviceTemplate = deviceLink.Spec.Template

	var deviceAnnotations = map[string]string{
		"edge.cattle.io/node-name":    deviceLink.Status.NodeName,
		"edge.cattle.io/adaptor-name": deviceLink.Status.AdaptorName,
	}
	var deviceLabels = collection.StringMapCopy(deviceTemplate.Labels)
	var deviceSpec = make(map[string]interface{}, 0)
	if deviceTemplate.Spec != nil {
		// NB(thxCode) apiserver will take care the format of `template.spec.raw`,
		// so we can consider it as a good JSON format.
		converter.TryUnmarshalJSON(deviceTemplate.Spec.Raw, &deviceSpec)
	}

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
