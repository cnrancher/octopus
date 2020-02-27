package predicate

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestModelChangedFuncs_GenericFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.GenericEvent
		expect bool
	}{
		{
			name: "without Meta",
			given: event.GenericEvent{
				Object: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "without Object",
			given: event.GenericEvent{
				Meta: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "none CRD instance",
			given: event.GenericEvent{
				Meta:   &edgev1alpha1.DeviceLink{},
				Object: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "CRD instance",
			given: event.GenericEvent{
				Meta:   &apiextensionsv1.CustomResourceDefinition{},
				Object: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = ModelChangedFuncs.Generic(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestModelChangedFuncs_UpdateFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.UpdateEvent
		expect bool
	}{
		{
			name: "without MetaOld",
			given: event.UpdateEvent{
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "without ObjectOld",
			given: event.UpdateEvent{
				MetaOld: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "none CRD instance",
			given: event.UpdateEvent{
				MetaOld:   &edgev1alpha1.DeviceLink{},
				ObjectOld: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "CRD instance",
			given: event.UpdateEvent{
				MetaOld:   &apiextensionsv1.CustomResourceDefinition{},
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "deleted CRD instance",
			given: event.UpdateEvent{
				MetaOld: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		var ret = ModelChangedFuncs.Update(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestModelChangedFuncs_DeleteFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.DeleteEvent
		expect bool
	}{
		{
			name: "without Meta",
			given: event.DeleteEvent{
				Object: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "without Object",
			given: event.DeleteEvent{
				Meta: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: false,
		},
		{
			name: "none CRD instance",
			given: event.DeleteEvent{
				Meta:   &edgev1alpha1.DeviceLink{},
				Object: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "CRD instance",
			given: event.DeleteEvent{
				Meta:   &apiextensionsv1.CustomResourceDefinition{},
				Object: &apiextensionsv1.CustomResourceDefinition{},
			},
			expect: true,
		},
		{
			name: "deleted CRD instance",
			given: event.DeleteEvent{
				Meta: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
				Object: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = ModelChangedFuncs.Delete(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
