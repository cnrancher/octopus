package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupResourceDeviceLink is group resource represented to the DeviceLink
	GroupResourceDeviceLink = schema.GroupResource{Group: GroupVersion.Group, Resource: "DeviceLink"}
)
