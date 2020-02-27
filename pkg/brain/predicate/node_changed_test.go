package predicate

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestNodeChangedFuncs_GenericFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.GenericEvent
		expect bool
	}{
		{
			name: "without Meta",
			given: event.GenericEvent{
				Object: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "without Object",
			given: event.GenericEvent{
				Meta: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "none Node instance",
			given: event.GenericEvent{
				Meta:   &edgev1alpha1.DeviceLink{},
				Object: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "Node instance",
			given: event.GenericEvent{
				Meta:   &corev1.Node{},
				Object: &corev1.Node{},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = NodeChangedFuncs.Generic(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestNodeChangedFuncs_UpdateFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.UpdateEvent
		expect bool
	}{
		{
			name: "without MetaOld",
			given: event.UpdateEvent{
				ObjectOld: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "without ObjectOld",
			given: event.UpdateEvent{
				MetaOld: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "none Node instance",
			given: event.UpdateEvent{
				MetaOld:   &edgev1alpha1.DeviceLink{},
				ObjectOld: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "Node instance",
			given: event.UpdateEvent{
				MetaOld:   &corev1.Node{},
				ObjectOld: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "deleted Node instance",
			given: event.UpdateEvent{
				MetaOld: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
				ObjectOld: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		var ret = NodeChangedFuncs.Update(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestNodeChangedFuncs_DeleteFunc(t *testing.T) {
	var testCases = []struct {
		name   string
		given  event.DeleteEvent
		expect bool
	}{
		{
			name: "without Meta",
			given: event.DeleteEvent{
				Object: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "without Object",
			given: event.DeleteEvent{
				Meta: &corev1.Node{},
			},
			expect: false,
		},
		{
			name: "none Node instance",
			given: event.DeleteEvent{
				Meta:   &edgev1alpha1.DeviceLink{},
				Object: &edgev1alpha1.DeviceLink{},
			},
			expect: true,
		},
		{
			name: "Node instance",
			given: event.DeleteEvent{
				Meta:   &corev1.Node{},
				Object: &corev1.Node{},
			},
			expect: true,
		},
		{
			name: "deleted Node instance",
			given: event.DeleteEvent{
				Meta: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
				Object: &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = NodeChangedFuncs.Delete(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
