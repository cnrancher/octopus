package v1alpha1

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var deviceLinkWebhookLog = ctrl.Log.WithName("webhook").WithName("deviceLink")

func (in *DeviceLink) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-edge-cattle-io-v1alpha1-devicelink,mutating=true,failurePolicy=fail,groups=edge.cattle.io,resources=devicelinks,verbs=create;update,versions=v1alpha1,name=mdevicelinks.edge.cattle.io

var _ webhook.Defaulter = &DeviceLink{}

func (in *DeviceLink) Default() {
	deviceLinkWebhookLog.V(4).Info("default", "object", fmt.Sprintf("%s/%s", in.GetNamespace(), in.GetName()))

	// fill `status.conditions` if it is empty
	if len(in.Status.Conditions) == 0 {
		in.Status.Conditions = []DeviceLinkCondition{
			{
				Type:           DeviceLinkNodeExisted,
				Status:         metav1.ConditionUnknown,
				LastUpdateTime: metav1.Time{Time: time.Now()},
				Reason:         "Confirming",
				Message:        "verify if there is a suitable node to schedule",
			},
		}
	}
}

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-edge-cattle-io-v1alpha1-devicelink,mutating=false,failurePolicy=fail,groups=edge.cattle.io,resources=devicelinks,versions=v1alpha1,name=vdevicelinks.edge.cattle.io

var _ webhook.Validator = &DeviceLink{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateCreate() error {
	deviceLinkWebhookLog.V(4).Info("validate create", "object", fmt.Sprintf("%s/%s", in.GetNamespace(), in.GetName()))

	var spec = in.Spec

	if spec.Model.APIVersion == "" {
		return errors.New("'spec.model.apiVersion' field could not be empty")
	} else if spec.Model.Kind == "" {
		return errors.New("'spec.model.kind' field could not be empty")
	}

	if spec.Adaptor.Node == "" {
		return errors.New("'spec.adaptor.node' field could not be blank")
	}

	if spec.Adaptor.Name == "" {
		return errors.New("'spec.adaptor.name' field could not be blank")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateUpdate(obj runtime.Object) error {
	deviceLinkWebhookLog.V(4).Info("validate update", "object", fmt.Sprintf("%s/%s", in.GetNamespace(), in.GetName()))

	var newSpec = in.Spec
	var oldSpec = (obj.(*DeviceLink)).Spec

	var newModelGVK = newSpec.Model.GroupVersionKind()
	var oldModelGVK = oldSpec.Model.GroupVersionKind()
	if newModelGVK.Kind != oldModelGVK.Kind {
		return errors.New("'spec.model.kind' field could not be modified")
	} else if newModelGVK.Group != newModelGVK.Group {
		return errors.New("the group of 'spec.model.apiVersion' field could not be modified")
	}

	if newSpec.Adaptor.Node != oldSpec.Adaptor.Node {
		return errors.New("'spec.adaptor.node' field could not be modified")
	}

	if newSpec.Adaptor.Name != oldSpec.Adaptor.Name {
		return errors.New("'spec.adaptor.name' field could not be modified")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateDelete() error {
	// nothing to do
	return nil
}
