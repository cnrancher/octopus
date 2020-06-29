package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByAdaptorField = "deviceLinkByAdaptor"

var deviceLinkByAdaptorIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByAdaptorField)

func DeviceLinkByAdaptorFuncFactory(nodeName string) func(runtime.Object) []string {
	return func(rawObj runtime.Object) []string {
		var link = object.ToDeviceLinkObject(rawObj)
		if link == nil {
			return nil
		}

		if link.Status.NodeName != nodeName {
			return nil
		}

		var adaptorName = link.Spec.Adaptor.Name
		if adaptorName != "" {
			deviceLinkByAdaptorIndexLog.V(6).Info("Indexed", "adaptorName", adaptorName, "object", object.GetNamespacedName(link))
			return []string{adaptorName}
		}
		return nil
	}
}
