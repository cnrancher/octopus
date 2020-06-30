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
	var targetNode = "edge-worker"
	var nonTargetNode = "edge-worker1"

	var testCases = []struct {
		name     string
		given    event.UpdateEvent
		expected bool
	}{
		{
			name: "without old object",
			given: generateUpdateEvent(
				nil,
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
			),
			expected: false,
		},
		{
			name: "non-DeviceLink instance",
			given: generateUpdateEvent(
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
			),
			expected: true,
		},
		{
			name: "different generation",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: targetNode,
							Name: "adaptors.edge.cattle.io/dummy",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: targetNode,
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: targetNode,
							Name: "adaptors.edge.cattle.io/dummy1",
						},
						Model: metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: targetNode,
						Model: &metav1.TypeMeta{
							Kind:       "DummySpecialDevice",
							APIVersion: "devices.edge.cattle.io/v1alpha1",
						},
					},
				},
			),
			expected: false,
		},
		{
			name: "different generation and node name",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: targetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: targetNode,
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 2,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Node: nonTargetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						NodeName: targetNode,
					},
				},
			),
			expected: true,
		},
		{
			name: "same generation",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
				},
			),
			expected: false,
		},
		{
			name: "same generation but different conditions",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
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
				&edgev1alpha1.DeviceLink{
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
			),
			expected: true,
		},
	}

	var predication = DeviceLinkChangedPredicate{}
	for _, tc := range testCases {
		var actual = predication.Update(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}
