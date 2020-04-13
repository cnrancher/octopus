package devicelink

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func GetNodeExistedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkNodeExisted).Status
}

func FailOnNodeExisted(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionFalse, "NotFound", message)
}

func SuccessOnNodeExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionTrue, "Found", "").
		next(edgev1alpha1.DeviceLinkModelExisted, "Confirming", "verify if there is a suitable model as a template")
}

func ToCheckNodeExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable node to schedule")
}

func GetModelExistedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkModelExisted).Status
}

func FailOnModelExisted(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionFalse, "NotFound", message)
}

func SuccessOnModelExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionTrue, "Found", "").
		next(edgev1alpha1.DeviceLinkAdaptorExisted, "Confirming", "verify if there is a suitable adaptor to access")
}

func ToCheckModelExisted(status *edgev1alpha1.DeviceLinkStatus) {

	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable model as a template")
}

func GetAdaptorExistedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkAdaptorExisted).Status
}

func FailOnAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionFalse, "NotFound", message)
}

func SuccessOnAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionTrue, "Found", "").
		next(edgev1alpha1.DeviceLinkDeviceCreated, "Creating", "verify if there is a corresponding device created")
}

func ToCheckAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionUnknown, "Confirming", "verify if there is a suitable adaptor to access")
}

func GetDeviceCreatedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkDeviceCreated).Status
}

func FailOnDeviceCreated(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceCreated, metav1.ConditionFalse, "Fail", message)
}

func SuccessOnDeviceCreated(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceCreated, metav1.ConditionTrue, "Success", "").
		next(edgev1alpha1.DeviceLinkDeviceConnected, "Connecting", "connect device")
}

func ToCheckDeviceCreated(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceCreated, metav1.ConditionUnknown, "Creating", "verify if there is a corresponding device created")
}

func GetDeviceConnectedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkDeviceConnected).Status
}

func FailOnDeviceConnected(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionFalse, "Unhealthy", message)
}

func SuccessOnDeviceConnected(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionTrue, "Healthy", "")
}

func ToCheckDeviceConnected(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionUnknown, "Connecting", "connect device")
}

type deviceLinkConditions []edgev1alpha1.DeviceLinkCondition

func (d deviceLinkConditions) get(t edgev1alpha1.DeviceLinkConditionType) edgev1alpha1.DeviceLinkCondition {
	for _, c := range d {
		if c.Type == t {
			return c
		}
	}
	return edgev1alpha1.DeviceLinkCondition{}
}

func (d deviceLinkConditions) did(t edgev1alpha1.DeviceLinkConditionType, result metav1.ConditionStatus, reason, message string) deviceLinkConditions {
	var (
		previous      []edgev1alpha1.DeviceLinkCondition
		lastItemLTT   = metav1.Time{Time: time.Now()}
		conditionsLen = len(d)
	)

	if conditionsLen > 0 {
		i := 0
		for _, c := range d {
			if c.Type != t {
				i++
			} else {
				break
			}
		}
		previous = d[:i]
		lastItemLTT = d[conditionsLen-1].LastUpdateTime
		// confirm the previous status
		if i > 1 {
			if d[i-1].Status != metav1.ConditionTrue {
				return previous
			}
		}
	}

	return append(previous,
		edgev1alpha1.DeviceLinkCondition{
			Type:               t,
			Status:             result,
			LastTransitionTime: lastItemLTT,
			LastUpdateTime:     metav1.Time{Time: time.Now()},
			Reason:             reason,
			Message:            message,
		},
	)
}

func (d deviceLinkConditions) next(t edgev1alpha1.DeviceLinkConditionType, reason, message string) deviceLinkConditions {
	var (
		lastItemLTT   = metav1.Time{Time: time.Now()}
		conditionsLen = len(d)
	)

	if conditionsLen != 0 {
		lastItemLTT = d[conditionsLen-1].LastUpdateTime
	}

	return append(d,
		edgev1alpha1.DeviceLinkCondition{
			Type:               t,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: lastItemLTT,
			LastUpdateTime:     lastItemLTT,
			Reason:             reason,
			Message:            message,
		},
	)
}
