package physical

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/rancher/octopus/adaptors/agent-device/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConstructDaemonSet(t *testing.T) {
	type given struct {
		name      string
		namespace string
		creator   string
		template  v1alpha1.PodTemplate
	}
	var testCases = []struct {
		given  given
		expect *appsv1.DaemonSet
	}{
		{
			given: given{
				name:      "nginx",
				namespace: "",
				creator:   "master",
				template: v1alpha1.PodTemplate{
					PodMetadata: v1alpha1.PodMetadata{Labels: map[string]string{"device": "test"}},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{Name: "nginx", Image: "nginx:1.14.2"}},
					},
				},
			},
			expect: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "DaemonSet",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx",
					Namespace: "",
					Annotations: map[string]string{
						AdaptorAnnotationName: AdaptorName,
					},
					Labels: map[string]string{
						GroupLabelName: "master",
					},
				},
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"device": "test"},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{{Name: "nginx", Image: "nginx:1.14.2"}},
						},
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"device": "test"},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		var ret = ConstructDaemonSet(tc.given.name, tc.given.namespace, tc.given.creator, tc.given.template)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
