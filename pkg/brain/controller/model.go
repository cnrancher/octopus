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
	"github.com/rancher/octopus/pkg/util/object"
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
	}

	if !object.IsActivating(&model) {
		// NB(thxCode) patches the CRD's name, as we don't use finalize to control the cleanup action.
		model.Name = req.Name

		// confirms Model isn't existed
		var links edgev1alpha1.DeviceLinkList
		if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
			log.Error(err, "Unable to list related DeviceLink of Model")
			return ctrl.Result{Requeue: true}, nil
		}
		for _, link := range links.Items {
			if link.GetModelExistedStatus() != metav1.ConditionTrue {
				continue
			}
			link.FailOnModelExisted("model isn't existed")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		log.V(5).Info("Model has been removed")
		return ctrl.Result{}, nil
	}

	// confirms Model is existed
	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByModelField: model.Name}); err != nil {
		log.Error(err, "Unable to list related DeviceLink of Model")
		return ctrl.Result{Requeue: true}, nil
	}
	for _, link := range links.Items {
		if link.GetModelExistedStatus() != metav1.ConditionFalse {
			continue
		}
		if !isModelAccepted(&link, &model) {
			link.FailOnModelExisted("model version isn't served")
		} else {
			link.SucceedOnModelExisted()
		}
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

func isModelAccepted(link *edgev1alpha1.DeviceLink, modelCRD *apiextensionsv1.CustomResourceDefinition) bool {
	var requestedVersion = link.Spec.Model.GroupVersionKind().Version
	for _, ver := range modelCRD.Spec.Versions {
		if ver.Name == requestedVersion {
			return ver.Served
		}
	}
	return false
}
