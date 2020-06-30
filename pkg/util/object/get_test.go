package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetNamespacedName(t *testing.T) {
	var testCases = []struct {
		name     string
		given    metav1.Object
		expected types.NamespacedName
	}{
		{
			name:     "nil instance",
			given:    nil,
			expected: types.NamespacedName{},
		},
		{
			name: "none namespaced instance",
			given: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expected: types.NamespacedName{
				Name: "test",
			},
		},
		{
			name: "namespaced instance",
			given: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
			},
			expected: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	for _, tc := range testCases {
		var actual = GetNamespacedName(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}
