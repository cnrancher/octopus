package status

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
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionFalse, "FinishedNodeValidation", message)
}

func SuccessOnNodeExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionTrue, "FinishedNodeValidation", "").
		next(edgev1alpha1.DeviceLinkModelExisted, "ValidatingModel", "verify if there is a suitable model existed")
}

func ToCheckNodeExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkNodeExisted, metav1.ConditionUnknown, "ValidatingNode", "verify if there is a suitable node existed")
}

func GetModelExistedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkModelExisted).Status
}

func FailOnModelExisted(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionFalse, "FinishedModelValidation", message)
}

func SuccessOnModelExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionTrue, "FinishedModelValidation", "").
		next(edgev1alpha1.DeviceLinkAdaptorExisted, "ValidatingAdaptor", "verify if there is a suitable adaptor existed")
}

func ToCheckModelExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkModelExisted, metav1.ConditionUnknown, "ValidatingModel", "verify if there is a suitable model existed")
}

func GetAdaptorExistedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkAdaptorExisted).Status
}

func FailOnAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionFalse, "FinishedAdaptorValidation", message)
}

func SuccessOnAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionTrue, "FinishedAdaptorValidation", "").
		next(edgev1alpha1.DeviceLinkDeviceConnected, "ValidatingDevice", "verify if there is a corresponding device ready")
}

func ToCheckAdaptorExisted(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkAdaptorExisted, metav1.ConditionUnknown, "ValidatingAdaptor", "verify if there is a suitable adaptor to access")
}

func GetDeviceConnectedStatus(status *edgev1alpha1.DeviceLinkStatus) metav1.ConditionStatus {
	return deviceLinkConditions(status.Conditions).get(edgev1alpha1.DeviceLinkDeviceConnected).Status
}

func FailOnDeviceConnected(status *edgev1alpha1.DeviceLinkStatus, message string) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionFalse, "FinishedDeviceConnected", message)
}

func SuccessOnDeviceConnected(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionTrue, "FinishedDeviceConnected", "")
}

func ToCheckDeviceConnected(status *edgev1alpha1.DeviceLinkStatus) {
	status.Conditions = deviceLinkConditions(status.Conditions).
		did(edgev1alpha1.DeviceLinkDeviceConnected, metav1.ConditionUnknown, "ConnectingDevice", "verify if there is a corresponding device connected")
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

	if conditionsLen != 0 {
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
