package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/brain/model"
	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByModelField = "DeviceLinkByModel"

var deviceLinkByModelIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByModelField)

func DeviceLinkByModelFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		deviceLinkByModelIndexLog.Error(nil, "received runtime object is not DeviceLink", "object", rawObj)
		return nil
	}

	var crdName = model.GetCRDNameOfGroupVersionKind(link.Status.Model.GroupVersionKind())
	if crdName != "" {
		return []string{crdName}
	}
	return nil
}
