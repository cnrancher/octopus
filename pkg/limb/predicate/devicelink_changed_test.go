package predicate

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
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
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
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
			name: "request the same node",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
			},
			expect: true,
		},
		{
			name: "requested the same node previously",
			given: event.UpdateEvent{
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker",
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
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
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
					},
				},
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
					},
					Spec: edgev1alpha1.DeviceLinkSpec{
						Adaptor: edgev1alpha1.DeviceAdaptor{
							Name: "adaptors.test.io/dummy",
							Node: "edge-worker1",
						},
					},
					Status: edgev1alpha1.DeviceLinkStatus{
						AdaptorName: "adaptors.test.io/dummy",
						NodeName:    "edge-worker1",
					},
				},
			},
			expect: false,
		},
	}

	var predication = DeviceLinkChangedPredicate{NodeName: "edge-worker"}
	for _, tc := range testCases {
		var ret = predication.Update(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
