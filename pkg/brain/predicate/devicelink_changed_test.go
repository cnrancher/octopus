package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkChangedPredicate_Update(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.UpdateEvent
		expect bool
	}{
		{
			name: "without MetaOld",
			given: event.UpdateEvent{
				MetaOld: nil,
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
			},
			expect: false,
		},
		{
			name: "non-DeviceLink instance",
			given: event.UpdateEvent{
				MetaOld: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				MetaNew: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				ObjectNew: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
			},
			expect: true,
		},
		{
			name: "different generation",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
							Name: "adaptors.edge.cattle.io/dummy",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
							Name: "adaptors.edge.cattle.io/dummy",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
							Name: "adaptors.edge.cattle.io/dummy1",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
							Name: "adaptors.edge.cattle.io/dummy1",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "different generation and node name",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: "test-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: "test-worker",
					},
				},
			},
			expect: true,
		},
		{
			name: "same generation",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
			},
			expect: false,
		},
		{
			name: "same generation and different conditions",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						Conditions: []edgev1alpha1.DeviceLinkCondition{
							{
								Type:   edgev1alpha1.DeviceLinkNodeExisted,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						Conditions: []edgev1alpha1.DeviceLinkCondition{
							{
								Type:   edgev1alpha1.DeviceLinkNodeExisted,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						Conditions: []edgev1alpha1.DeviceLinkCondition{
							{
								Type:   edgev1alpha1.DeviceLinkNodeExisted,
								Status: metav1.ConditionTrue,
							},
							{
								Type:   edgev1alpha1.DeviceLinkModelExisted,
								Status: metav1.ConditionUnknown,
							},
						},
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						Conditions: []edgev1alpha1.DeviceLinkCondition{
							{
								Type:   edgev1alpha1.DeviceLinkNodeExisted,
								Status: metav1.ConditionTrue,
							},
							{
								Type:   edgev1alpha1.DeviceLinkModelExisted,
								Status: metav1.ConditionUnknown,
							},
						},
					},
				},
			},
			expect: true,
		},
	}

	var predication = DeviceLinkChangedPredicate{}
	for _, tc := range testCases {
		var ret = predication.Update(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}
