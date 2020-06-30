package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (in *DeviceLink) FailOnNodeExisted(message string) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkNodeExisted, metav1.ConditionFalse, "NotFound", message, in.Status.NodeName != in.Spec.Adaptor.Node)
	in.Status.NodeName = ""
}

func (in *DeviceLink) SucceedOnNodeExisted(node *corev1.Node) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkNodeExisted, metav1.ConditionTrue, "Found", "", in.Status.NodeName != in.Spec.Adaptor.Node).
		next(DeviceLinkModelExisted, "Confirming", "verify if there is a suitable model as a template")
	in.Status.NodeName = in.Spec.Adaptor.Node
	if node != nil {
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalDNS:
				in.Status.NodeInternalDNS = address.Address
			case corev1.NodeInternalIP:
				in.Status.NodeInternalIP = address.Address
			case corev1.NodeExternalDNS:
				in.Status.NodeExternalDNS = address.Address
			case corev1.NodeExternalIP:
				in.Status.NodeExternalIP = address.Address
			case corev1.NodeHostName:
				in.Status.NodeHostName = address.Address
			}
		}
	}
}

func (in *DeviceLink) ToCheckNodeExisted() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkNodeExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable node to schedule", false)
}

func (in *DeviceLink) GetNodeExistedStatus() metav1.ConditionStatus {
	if in == nil {
		return ""
	}
	return deviceLinkConditions(in.Status.Conditions).get(DeviceLinkNodeExisted).Status
}

func (in *DeviceLink) FailOnModelExisted(message string) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkModelExisted, metav1.ConditionFalse, "NotFound", message, in.Status.Model == nil || *in.Status.Model != in.Spec.Model)
	in.Status.Model = nil
}

func (in *DeviceLink) SucceedOnModelExisted() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkModelExisted, metav1.ConditionTrue, "Found", "", in.Status.Model == nil || *in.Status.Model != in.Spec.Model).
		next(DeviceLinkAdaptorExisted, "Confirming", "verify if there is a suitable adaptor to access")
	in.Status.Model = &in.Spec.Model
}

func (in *DeviceLink) ToCheckModelExisted() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkModelExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable model as a template", false)
}

func (in *DeviceLink) GetModelExistedStatus() metav1.ConditionStatus {
	if in == nil {
		return ""
	}
	return deviceLinkConditions(in.Status.Conditions).get(DeviceLinkModelExisted).Status
}

func (in *DeviceLink) FailOnAdaptorExisted(message string) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkAdaptorExisted, metav1.ConditionFalse, "NotFound", message, in.Status.AdaptorName != in.Spec.Adaptor.Name)
	in.Status.AdaptorName = ""
}

func (in *DeviceLink) SucceedOnAdaptorExisted() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkAdaptorExisted, metav1.ConditionTrue, "Found", "", in.Status.AdaptorName != in.Spec.Adaptor.Name).
		next(DeviceLinkDeviceCreated, "Creating", "verify if there is a corresponding device created")
	in.Status.AdaptorName = in.Spec.Adaptor.Name
}

func (in *DeviceLink) ToCheckAdaptorExisted() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkAdaptorExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable adaptor to access", false)
}

func (in *DeviceLink) GetAdaptorExistedStatus() metav1.ConditionStatus {
	if in == nil {
		return ""
	}
	return deviceLinkConditions(in.Status.Conditions).get(DeviceLinkAdaptorExisted).Status
}

func (in *DeviceLink) FailOnDeviceCreated(message string) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceCreated, metav1.ConditionFalse, "Fail", message, false)
}

func (in *DeviceLink) SucceedOnDeviceCreated() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceCreated, metav1.ConditionTrue, "Success", "", false).
		next(DeviceLinkDeviceConnected, "Connecting", "connect device")
}

func (in *DeviceLink) ToCheckDeviceCreated() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceCreated, metav1.ConditionUnknown, "Creating", "verify if there is a corresponding device created", false)
}

func (in *DeviceLink) GetDeviceCreatedStatus() metav1.ConditionStatus {
	if in == nil {
		return ""
	}
	return deviceLinkConditions(in.Status.Conditions).get(DeviceLinkDeviceCreated).Status
}

func (in *DeviceLink) FailOnDeviceConnected(message string) {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceConnected, metav1.ConditionFalse, "Unhealthy", message, false)
}

func (in *DeviceLink) SucceedOnDeviceConnected() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceConnected, metav1.ConditionTrue, "Healthy", "", false)
}

func (in *DeviceLink) ToCheckDeviceConnected() {
	if in == nil {
		return
	}
	in.Status.Conditions = deviceLinkConditions(in.Status.Conditions).
		did(DeviceLinkDeviceConnected, metav1.ConditionUnknown, "Connecting", "connect device", false)
}

func (in *DeviceLink) GetDeviceConnectedStatus() metav1.ConditionStatus {
	if in == nil {
		return ""
	}
	return deviceLinkConditions(in.Status.Conditions).get(DeviceLinkDeviceConnected).Status
}

type deviceLinkConditions []DeviceLinkCondition

func (d deviceLinkConditions) get(t DeviceLinkConditionType) DeviceLinkCondition {
	for _, c := range d {
		if c.Type == t {
			return c
		}
	}
	return DeviceLinkCondition{}
}

func (d deviceLinkConditions) did(t DeviceLinkConditionType, result metav1.ConditionStatus, reason, message string, must bool) deviceLinkConditions {
	var (
		size               = len(d)
		lastTransitionTime = metav1.NewTime(time.Now())
		list               = make([]DeviceLinkCondition, 0, size)
	)

	if size > 0 {
		if !d[size-1].LastUpdateTime.IsZero() {
			lastTransitionTime = d[size-1].LastUpdateTime
		}

		for _, c := range d {
			if c.Type == t {
				if !must {
					// NB(thxCode) doesn't need to update the last condition if {status,reason,message} are the same.
					if c.Status == result && c.Reason == reason && c.Message == message {
						return d
					}
				}
				break
			}
			list = append(list, c)
		}
		if len(list) > 0 {
			// NB(thxCode) doesn't append new condition if the status of previous/penultimate is False/Unknown.
			if list[len(list)-1].Status != metav1.ConditionTrue {
				return list
			}
		}
	}

	return append(list,
		DeviceLinkCondition{
			Type:               t,
			Status:             result,
			LastTransitionTime: lastTransitionTime,
			LastUpdateTime:     metav1.NewTime(time.Now()),
			Reason:             reason,
			Message:            message,
		},
	)
}

func (d deviceLinkConditions) next(t DeviceLinkConditionType, reason, message string) deviceLinkConditions {
	var (
		size               = len(d)
		lastTransitionTime = metav1.NewTime(time.Now())
	)

	if size > 0 {
		if !d[size-1].LastUpdateTime.IsZero() {
			lastTransitionTime = d[size-1].LastUpdateTime
		}

		for _, c := range d {
			if c.Type == t {
				// NB(thxCode) doesn't need to update condition if it is in the list.
				return d
			}
		}
	}

	return append(d,
		DeviceLinkCondition{
			Type:               t,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: lastTransitionTime,
			LastUpdateTime:     lastTransitionTime,
			Reason:             reason,
			Message:            message,
		},
	)
}
