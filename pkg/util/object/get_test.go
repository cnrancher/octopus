package object

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetNamespacedName(t *testing.T) {
	var testCases = []struct {
		name   string
		given  metav1.Object
		expect types.NamespacedName
	}{
		{
			name:   "nil instance",
			given:  nil,
			expect: types.NamespacedName{},
		},
		{
			name: "none namespaced instance",
			given: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expect: types.NamespacedName{
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
			expect: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
		},
	}

	for _, tc := range testCases {
		var ret = GetNamespacedName(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
