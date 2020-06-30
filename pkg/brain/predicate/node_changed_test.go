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
		name     string
		given    event.UpdateEvent
		expected bool
	}{
		{
			name: "without old object",
			given: generateUpdateEvent(
				nil,
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "edge-worker",
					},
				},
			),
			expected: false,
		},
		{
			name: "non-Node instance",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
			),
			expected: true,
		},
		{
			name: "changed Node instance's addresses",
			given: generateUpdateEvent(
				&corev1.Node{
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
				&corev1.Node{
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
			),
			expected: true,
		},
	}

	var predication = NodeChangedPredicate{}
	for _, tc := range testCases {
		var actual = predication.Update(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}
