package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/model"
	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByModelField = "deviceLinkByModel"

var deviceLinkByModelIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByModelField)

func DeviceLinkByModelFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		return nil
	}

	var crdName = model.GetCRDNameOfGroupVersionKind(link.Status.Model.GroupVersionKind())
	if crdName != "" {
		deviceLinkByModelIndexLog.V(0).Info("Index DeviceLink by Model", "crdName", crdName, "object", object.GetNamespacedName(link))
		return []string{crdName}
	}
	return nil
}
