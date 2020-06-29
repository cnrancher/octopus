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
	var testNode = "edge-worker"
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
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
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
			name: "same generation",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
					},
				},
			},
			expect: false,
		},
		{
			name: "different generation and requested the same node",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
					},
				},
			},
			expect: true,
		},
		{ // this case is used for cancel the previous connection.
			name: "different generation but requested the same node previously",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode,
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
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
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode,
					},
				},
			},
			expect: true,
		},
		{
			name: "request another node",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
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
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode + "1",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:  "default",
						Name:       "test",
						Generation: 1,
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: testNode + "1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    testNode + "1",
					},
				},
			},
			expect: false,
		},
	}

	var predication = DeviceLinkChangedPredicate{NodeName: testNode}
	for _, tc := range testCases {
		var ret = predication.Update(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}
