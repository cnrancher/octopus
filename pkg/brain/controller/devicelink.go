package controller

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

// DeviceLinkReconciler reconciles a DeviceLink object
type DeviceLinkReconciler struct {
	client.Client
	record.EventRecorder

	Log logr.Logger
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

func (r *DeviceLinkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = context.Background()
	var log = r.Log.WithValues("devicelink", req.NamespacedName)

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

	if object.IsDeleted(&link) {
		return ctrl.Result{}, nil
	}

	// validates node existing or not
	switch devicelink.GetNodeExistedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if link.Spec.Adaptor.Node != link.Status.Adaptor.Node {
			devicelink.ToCheckNodeExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if link.Spec.Adaptor.Node != link.Status.Adaptor.Node {
			devicelink.ToCheckNodeExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
	default:
		var node corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: link.Spec.Adaptor.Node}, &node); err != nil {
			if !apierrs.IsNotFound(err) {
				log.Error(err, "unable to fetch the adaptor node of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		if object.IsActivating(&node) {
			devicelink.SuccessOnNodeExisted(&link.Status)
			r.Eventf(&link, "Normal", "Validated", "found the adaptor node")
		} else {
			devicelink.FailOnNodeExisted(&link.Status, "adaptor node isn't existed")
			r.Eventf(&link, "Warning", "FailedValidate", "could not find the adaptor node")
		}

		link.Status.Adaptor.Node = link.Spec.Adaptor.Node
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// validates model existing or not
	switch devicelink.GetModelExistedStatus(&link.Status) {
	case metav1.ConditionFalse:
		if link.Spec.Model != link.Status.Model {
			devicelink.ToCheckModelExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	case metav1.ConditionTrue:
		if link.Spec.Model != link.Status.Model {
			devicelink.ToCheckModelExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, nil
		}
	default:
		var mGVK = link.Spec.Model.GroupVersionKind()
		var m = apiextensionsv1.CustomResourceDefinition{}
		if err := r.Get(
			ctx,
			types.NamespacedName{Name: model.GetCRDNameOfGroupVersionKind(mGVK)},
			&m,
		); err != nil {
			if !apierrs.IsNotFound(err) {
				log.Error(err, "unable to fetch the model of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		// delete the stale device
		var device = model.NewInstanceOfTypeMeta(link.Spec.Model)
		if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
			if !apierrs.IsNotFound(err) && !meta.IsNoMatchError(err) {
				log.Error(err, "unable to fetch the stale device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}
		if object.IsActivating(&device) {
			if err := r.Delete(ctx, &device); err != nil {
				log.Error(err, "unable to delete the stale device of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		link.Status.Model = link.Spec.Model
		if object.IsActivating(&m) {
			var versionServed bool
			for _, ver := range m.Spec.Versions {
				if ver.Name == mGVK.Version {
					versionServed = ver.Served
					break
				}
			}

			if versionServed {
				devicelink.SuccessOnModelExisted(&link.Status)
				r.Eventf(&link, "Normal", "Validated", "found the model")
			} else {
				devicelink.FailOnModelExisted(&link.Status, "model version isn't served")
				r.Eventf(&link, "Warning", "FailedValidate", "could not find the version of model")
			}
		} else {
			devicelink.FailOnModelExisted(&link.Status, "model isn't existed")
			r.Eventf(&link, "Warning", "FailedValidate", "could not find the model")
		}
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *DeviceLinkReconciler) SetupWithManager(name string, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(name + ".DeviceLink").
		For(&edgev1alpha1.DeviceLink{}).
		Complete(r)
}
