package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByAdaptorField = "DeviceLinkByAdaptor"

var deviceLinkByAdaptorIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByAdaptorField)

func DeviceLinkByAdaptorFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		deviceLinkByAdaptorIndexLog.Error(nil, "received runtime object is not DeviceLink", "object", rawObj)
		return nil
	}

	var name = link.Status.Adaptor.Name
	if name != "" {
		return []string{name}
	}
	return nil
}
