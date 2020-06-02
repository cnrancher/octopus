package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByAdaptorField = "deviceLinkByAdaptor"

var deviceLinkByAdaptorIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByAdaptorField)

func DeviceLinkByAdaptorFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		return nil
	}

	var adaptorName = link.Status.AdaptorName
	if adaptorName != "" {
		deviceLinkByAdaptorIndexLog.V(0).Info("Index DeviceLink by Adaptor", "adaptorName", adaptorName, "object", object.GetNamespacedName(link))
		return []string{adaptorName}
	}
	return nil
}
