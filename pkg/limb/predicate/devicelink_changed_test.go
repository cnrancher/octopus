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
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
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
			name: "same generation",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    targetNode,
					},
				},
			),
			expected: false,
		},
		{
			name: "different generation and requested the same node",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    targetNode,
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
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    targetNode,
					},
				},
			),
			expected: true,
		},
		{ // this case is used for cancel the previous connection.
			name: "different generation but requested the same node previously",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: targetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    targetNode,
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
							Name: "adaptors.test.io/dummy",
							Node: nonTargetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    targetNode,
					},
				},
			),
			expected: true,
		},
		{
			name: "request another node",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: nonTargetNode,
						},
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: nonTargetNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    nonTargetNode,
					},
				},
			),
			expected: false,
		},
	}

	var predication = DeviceLinkChangedPredicate{NodeName: targetNode}
	for _, tc := range testCases {
		var actual = predication.Update(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}
