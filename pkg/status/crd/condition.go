package crd

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetEstablished(status *apiextensionsv1.CustomResourceDefinitionStatus) metav1.ConditionStatus {
	return unify(crdConditions(status.Conditions).get(apiextensionsv1.Established).Status)
}

type crdConditions []apiextensionsv1.CustomResourceDefinitionCondition

func (d crdConditions) get(t apiextensionsv1.CustomResourceDefinitionConditionType) apiextensionsv1.CustomResourceDefinitionCondition {
	for _, c := range d {
		if c.Type == t {
			return c
		}
	}
	return apiextensionsv1.CustomResourceDefinitionCondition{}
}

func unify(status apiextensionsv1.ConditionStatus) metav1.ConditionStatus {
	return metav1.ConditionStatus(status)
}
