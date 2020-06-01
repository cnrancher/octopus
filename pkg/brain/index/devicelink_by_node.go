package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByNodeField = "deviceLinkByNode"

var deviceLinkByNodeIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByNodeField)

func DeviceLinkByNodeFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		return nil
	}

	var nodeName = link.Status.NodeName
	if nodeName != "" {
		deviceLinkByNodeIndexLog.V(0).Info("Index DeviceLink by Node", "nodeName", nodeName, "object", object.GetNamespacedName(link))
		return []string{nodeName}
	}
	return nil
}
