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
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/util/collection"
	"github.com/rancher/octopus/pkg/util/object"
)

const (
	ReconcilingNode = "edge.cattle.io/octopus-brain"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client

	Log logr.Logger
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var ctx = context.Background()
	var log = r.Log.WithValues("node", req.NamespacedName)

	// fetch node
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch Node")
			return ctrl.Result{Requeue: true}, nil
		}
		// ignores error, since they can't be fixed by an immediate requeue
		return ctrl.Result{}, nil
	}

	if object.IsDeleted(&node) {
		if !collection.StringSliceContain(node.Finalizers, ReconcilingNode) {
			return ctrl.Result{}, nil
		}

		// move link NodeExisted condition to `False`
		var links edgev1alpha1.DeviceLinkList
		if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByNodeField: node.Name}); err != nil {
			log.Error(err, "unable to list related DeviceLink of Node")
			return ctrl.Result{Requeue: true}, nil
		}
		for _, link := range links.Items {
			if devicelink.GetNodeExistedStatus(&link.Status) == metav1.ConditionFalse {
				continue
			}
			devicelink.FailOnNodeExisted(&link.Status, "adaptor node isn't existed")
			if err := r.Status().Update(ctx, &link); err != nil {
				log.Error(err, "unable to change the status of DeviceLink")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		// remove finalizer
		node.Finalizers = collection.StringSliceRemove(node.Finalizers, ReconcilingNode)
		if err := r.Update(ctx, &node); err != nil {
			log.Error(err, "unable to remove finalizer from Node")
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	}

	// add finalizer if needed
	if !collection.StringSliceContain(node.Finalizers, ReconcilingNode) {
		node.Finalizers = append(node.Finalizers, ReconcilingNode)
		if err := r.Update(ctx, &node); err != nil {
			log.Error(err, "unable to add finalizer to Node")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	// move link NodeExisted condition from `False` to `True`
	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByNodeField: node.Name}); err != nil {
		log.Error(err, "unable to list related DeviceLink of Node")
		return ctrl.Result{Requeue: true}, nil
	}
	for _, link := range links.Items {
		if devicelink.GetNodeExistedStatus(&link.Status) != metav1.ConditionFalse {
			continue
		}
		devicelink.ToCheckNodeExisted(&link.Status)
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) SetupWithManager(name string, mgr ctrl.Manager) error {
	// indexes DeviceLink by `spec.adaptor.node`
	if err := mgr.GetFieldIndexer().IndexField(
		&edgev1alpha1.DeviceLink{},
		index.DeviceLinkByNodeField,
		index.DeviceLinkByNodeFunc,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name + ".Node").
		For(&corev1.Node{}).
		WithEventFilter(predicate.NodeChangedFuncs).
		Complete(r)
}
