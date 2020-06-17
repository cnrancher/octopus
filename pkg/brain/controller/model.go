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
	"github.com/rancher/octopus/pkg/status/crd"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
)

const (
	ReconcilingModel = "edge.cattle.io/octopus-brain"
)

// ModelReconciler reconciles a CRD object
type ModelReconciler struct {
	client.Client

	Ctx context.Context
	Log logr.Logger
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;update;patch

func (r *ModelReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = r.Ctx
	var log = r.Log.WithValues("crd", req.NamespacedName)

	// fetches model
	var model apiextensionsv1.CustomResourceDefinition
	if err := r.Get(ctx, req.NamespacedName, &model); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch Model")
			return ctrl.Result{Requeue: true}, nil
		}
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}

	if object.IsDeleted(&model) {
		if !collection.StringSliceContain(model.Finalizers, ReconcilingModel) {
			return ctrl.Result{}, nil
		}

		// moves link ModelExisted condition from `True` to `Unknown`
		var links edgev1alpha1.DeviceLinkList
		if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
			log.Error(err, "Unable to list related DeviceLink of Model")
			return ctrl.Result{Requeue: true}, nil
		}
		for _, link := range links.Items {
			if devicelink.GetModelExistedStatus(&link.Status) != metav1.ConditionTrue {
				continue
			}
			devicelink.ToCheckModelExisted(&link.Status)
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		// removes finalizer
		model.Finalizers = collection.StringSliceRemove(model.Finalizers, ReconcilingModel)
		if err := r.Update(ctx, &model); err != nil {
			log.Error(err, "Unable to remove finalizer from Model")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// adds finalizer if needed
	if !collection.StringSliceContain(model.Finalizers, ReconcilingModel) {
		if crd.GetEstablished(&model.Status) != metav1.ConditionTrue {
			return ctrl.Result{Requeue: true}, nil
		}
		model.Finalizers = append(model.Finalizers, ReconcilingModel)
		if err := r.Update(ctx, &model); err != nil {
			log.Error(err, "Unable to add finalizer to CRD")
			return ctrl.Result{Requeue: true}, nil
		}
		// NB(thxCode) keeps going down, no need to reconcile again:
		//     `return ctrl.Result{}, nil`,
		// the predication will prevent the updated reconciling.
	}

	// moves link ModelExisted condition from `False` to `True`
	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
		log.Error(err, "Unable to list related DeviceLink of Model")
		return ctrl.Result{Requeue: true}, nil
	}
	for _, link := range links.Items {
		if devicelink.GetModelExistedStatus(&link.Status) != metav1.ConditionFalse {
			continue
		}
		devicelink.ToCheckModelExisted(&link.Status)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// indexes DeviceLink by `status.model`
	if err := mgr.GetFieldIndexer().IndexField(
		r.Ctx,
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByModelField,
		index.DeviceLinkByModelFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("brain_crd").
		For(&apiextensionsv1.CustomResourceDefinition{}).
		WithEventFilter(predicate.ModelChangedPredicate{}).
		Complete(r)
}
