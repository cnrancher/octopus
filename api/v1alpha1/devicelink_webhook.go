package v1alpha1

import (
	"time"

	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var deviceWebhookLog = ctrl.Log.WithName("webhook").WithName("DeviceLink")

func (in *DeviceLink) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-edge-cattle-io-v1alpha1-device,mutating=true,failurePolicy=fail,groups=edge.cattle.io,resources=devicelinks,verbs=create;update,versions=v1alpha1,name=mdevicelinks.edge.cattle.io

var _ webhook.Defaulter = &DeviceLink{}

func (in *DeviceLink) Default() {
	deviceWebhookLog.V(0).Info("default", "name", in.Name)

	// only allow to specify the `labels` and `annotations` of `spec.template.metadata`
	var template = in.Spec.Template
	var srcTemplateObjectMeta = template.ObjectMeta.DeepCopy()
	template.ObjectMeta = metav1.ObjectMeta{
		Labels:      srcTemplateObjectMeta.Labels,
		Annotations: srcTemplateObjectMeta.Annotations,
	}

	// fill `status.conditions` if it is empty
	if len(in.Status.Conditions) == 0 {
		in.Status.Conditions = []DeviceLinkCondition{
			{
				Type:           DeviceLinkNodeExisted,
				Status:         metav1.ConditionUnknown,
				LastUpdateTime: metav1.Time{Time: time.Now()},
				Reason:         "ValidatingNode",
				Message:        "verify if there is a suitable node to schedule",
			},
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-edge-cattle-io-v1alpha1-device,mutating=false,failurePolicy=fail,groups=edge.cattle.io,resources=devicelinks,versions=v1alpha1,name=vdevicelinks.edge.cattle.io

var _ webhook.Validator = &DeviceLink{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateCreate() error {
	deviceWebhookLog.V(0).Info("validate create", "name", in.Name)

	var spec = in.Spec

	if spec.Model.APIVersion == "" || spec.Model.Kind == "" {
		return apierrs.NewForbidden(
			GroupResourceDeviceLink,
			"spec.model",
			errors.New("field could not be empty"),
		)
	}

	if spec.Adaptor.Node == "" {
		return apierrs.NewForbidden(
			GroupResourceDeviceLink,
			"spec.adaptor.node",
			errors.New("field could not be blank"),
		)
	}

	if spec.Adaptor.Name == "" {
		return apierrs.NewForbidden(
			GroupResourceDeviceLink,
			"spec.adaptor.name",
			errors.New("field could not be blank"),
		)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateUpdate(_ runtime.Object) error {
	// nothing to do
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *DeviceLink) ValidateDelete() error {
	// nothing to do
	return nil
}
