package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/limb/index"
	"github.com/rancher/octopus/pkg/status/devicelink"
	"github.com/rancher/octopus/pkg/suctioncup"
	"github.com/rancher/octopus/pkg/util/log/handler"
)

// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks,verbs=list
// +kubebuilder:rbac:groups=edge.cattle.io,resources=devicelinks/status,verbs=get;update;patch

func (r *DeviceLinkReconciler) ReceiveAdaptorStatus(req suctioncup.RequestAdaptorStatus) (suctioncup.Response, error) {
	var ctx = context.Background()
	var log = r.Log.WithName("ReceiveAdaptorStatus").WithValues("adaptor", req.Name)

	defer runtime.HandleCrash(handler.NewPanicsLogHandler(log))

	var links edgev1alpha1.DeviceLinkList
	if err := r.List(ctx, &links, client.MatchingFields{index.DeviceLinkByAdaptorField: req.Name}); err != nil {
		log.Error(err, "unable to list related DeviceLink of adaptor")
		return suctioncup.Response{Requeue: true}, nil
	}

	if req.Registered {
		log.Info("adaptor is registered")
	} else {
		log.Info("adaptor is unregistered")
	}

	for _, link := range links.Items {
		// filter out the corresponding links
		if link.Spec.Adaptor.Node != r.NodeName {
			continue
		}

		if req.Registered {
			if devicelink.GetAdaptorExistedStatus(&link.Status) == metav1.ConditionFalse {
				devicelink.SuccessOnAdaptorExisted(&link.Status)
				r.Eventf(&link, "Normal", "Validated", "found the adaptor")
			}
		} else {
			if devicelink.GetAdaptorExistedStatus(&link.Status) != metav1.ConditionFalse {
				devicelink.FailOnAdaptorExisted(&link.Status, "the adaptor is unregistered")
				r.Eventf(&link, "Warning", "FailedValidate", "the adaptor is unregistered")
			}
		}
		if err := r.Status().Update(ctx, &link); err != nil {
			log.Error(err, "unable to change the status of DeviceLink")
			return suctioncup.Response{Requeue: true}, nil
		}
	}

	return suctioncup.Response{}, nil
}
