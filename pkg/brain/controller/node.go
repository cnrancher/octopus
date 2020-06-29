package controller

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/brain/index"
	"github.com/rancher/octopus/pkg/brain/predicate"
	"github.com/rancher/octopus/pkg/util/object"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client

	Ctx context.Context
	Log logr.Logger
}

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch

func (r *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = r.Ctx
	var log = r.Log.WithValues("node", req.NamespacedName)

	// fetches node
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch Node")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	if !object.IsActivating(&node) {
		// NB(thxCode) patches the Node's name, as we don't use finalize to control the cleanup action.
		node.Name = req.Name

		// confirms Node isn't existed
		var links edgev1alpha1.DeviceLinkList
		if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByNodeField: node.Name}); err != nil {
			log.Error(err, "Unable to list related DeviceLink of Node")
			return ctrl.Result{Requeue: true}, nil
		}
		for _, link := range links.Items {
			if link.GetNodeExistedStatus() != metav1.ConditionTrue {
				continue
			}
			link.FailOnNodeExisted("adaptor node isn't existed")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "Unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		log.V(5).Info("Node has been removed")
		return ctrl.Result{}, nil
	}

	// confirms Node is existed
	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByNodeField: node.Name}); err != nil {
		log.Error(err, "Unable to list related DeviceLink of Node")
		return ctrl.Result{Requeue: true}, nil
	}
	for _, link := range links.Items {
		if link.GetNodeExistedStatus() != metav1.ConditionFalse {
			continue
		}
		link.SucceedOnNodeExisted(&node)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "Unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		r.Ctx,
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByNodeField,
		index.DeviceLinkByNodeFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("brain_node").
		For(&corev1.Node{}).
		WithEventFilter(predicate.NodeChangedPredicate{}).
		Complete(r)
}
