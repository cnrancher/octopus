package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestNodeChangedPredicate_Update(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.UpdateEvent
		expect bool
	}{
		{
			name: "without MetaOld",
			given: event.UpdateEvent{
				MetaOld: nil,
				ObjectOld: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
				},
				MetaNew: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
				},
				ObjectNew: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
				},
			},
			expect: false,
		},
		{
			name: "non-Node instance",
			given: event.UpdateEvent{
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
			},
			expect: true,
		},
		{
			name: "changed Node instance's addresses",
			given: event.UpdateEvent{
				MetaOld: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "172.18.0.3",
							},
							{
								Type:    corev1.NodeHostName,
								Address: "edge-worker",
							},
						},
					},
				},
				ObjectOld: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "172.18.0.3",
							},
							{
								Type:    corev1.NodeHostName,
								Address: "edge-worker",
							},
						},
					},
				},
				MetaNew: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "172.18.0.3",
							},
							{
								Type:    corev1.NodeHostName,
								Address: "edge-worker",
							},
							{
								Type:    corev1.NodeInternalDNS,
								Address: "edge-worker.octopus.internal",
							},
						},
					},
				},
				ObjectNew: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "172.18.0.3",
							},
							{
								Type:    corev1.NodeHostName,
								Address: "edge-worker",
							},
							{
								Type:    corev1.NodeInternalDNS,
								Address: "edge-worker.octopus.internal",
							},
						},
					},
				},
			},
			expect: true,
		},
	}

	var predication = NodeChangedPredicate{}
	for _, tc := range testCases {
		var ret = predication.Update(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}
