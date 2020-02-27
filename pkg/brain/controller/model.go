package controller

import (
	"context"

	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/brain/index"
	"github.com/rancher/octopus/pkg/brain/predicate"
	"github.com/rancher/octopus/pkg/status"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
)

const (
	ReconcilingModel = "edge.cattle.io/octopus-brain"
)

// ModelReconciler reconciles a Node object
type ModelReconciler struct {
	client.Client

	Log logr.Logger
}

// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *ModelReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("crd", req.NamespacedName)

	defer func() {
		log.V(0).Info("reconcile out")
	}()
	log.V(0).Info("reconcile in")

	// fetch model
	var model apiextensionsv1.CustomResourceDefinition
	if err := r.Get(ctx, req.NamespacedName, &model); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch Model")
			return ctrl.Result{Requeue: true}, nil
		}
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}

	if object.IsDeleted(&model) {
		if !collection.StringSliceContain(model.Finalizers, ReconcilingModel) {
			return ctrl.Result{}, nil
		}

		// move link NodeExisted condition to `False`
		var links edgev1alpha1.DeviceLinkList
		if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
			log.Error(err, "unable to list related DeviceLink of Model")
			return ctrl.Result{Requeue: true}, nil
		}
		for _, link := range links.Items {
			if status.GetNodeExistedStatus(&link.Status) == metav1.ConditionFalse {
				continue
			}
			status.FailOnModelExisted(&link.Status, "model isn't existed")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		// remove finalizer
		model.Finalizers = collection.StringSliceRemove(model.Finalizers, ReconcilingModel)
		if err := r.Update(ctx, &model); err != nil {
			log.Error(err, "unable to remove finalizer from Model")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// add finalizer if needed
	if !collection.StringSliceContain(model.Finalizers, ReconcilingModel) {
		model.Finalizers = append(model.Finalizers, ReconcilingModel)
		if err := r.Update(ctx, &model); err != nil {
			log.Error(err, "unable to add finalizer to Node")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// move link ModelExisted condition from `False` to `True`
	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
		log.Error(err, "unable to list related DeviceLink of Model")
		return ctrl.Result{Requeue: true}, nil
	}
	for _, link := range links.Items {
		if status.GetModelExistedStatus(&link.Status) != metav1.ConditionFalse {
			continue
		}
		status.ToCheckModelExisted(&link.Status)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) SetupWithManager(name string, mgr ctrl.Manager) error {
	// indexes DeviceLink by `spec.model`
	if err := mgr.GetFieldIndexer().IndexField(
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByModelField,
		index.DeviceLinkByModelFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name + ".CustomResourceDefinition").
		For(&apiextensionsv1.CustomResourceDefinition{}).
		WithEventFilter(predicate.ModelChangedFuncs).
		Complete(r)
}
