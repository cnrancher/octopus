package index

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/object"
)

const DeviceLinkByNodeField = "DeviceLinkByNode"

var deviceLinkByNodeIndexLog = ctrl.Log.WithName("index").WithName(DeviceLinkByNodeField)

func DeviceLinkByNodeFunc(rawObj runtime.Object) []string {
	var link = object.ToDeviceLinkObject(rawObj)
	if link == nil {
		deviceLinkByNodeIndexLog.Error(nil, "received runtime object is not DeviceLink", "object", rawObj)
		return nil
	}

	var nodeName = link.Status.Adaptor.Node
	if nodeName != "" {
		return []string{nodeName}
	}
	return nil
}
