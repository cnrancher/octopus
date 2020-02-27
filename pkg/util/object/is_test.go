package object

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestIsDeleted(t *testing.T) {
	var testCases = []struct {
		name   string
		given  metav1.Object
		expect bool
	}{
		{
			name: "alive instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expect: false,
		},
		{
			name: "deleted instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "default",
					Name:              "test2",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		var ret = IsDeleted(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestIsZero(t *testing.T) {
	var testCases = []struct {
		name   string
		given  metav1.Object
		expect bool
	}{
		{
			name: "existing instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expect: false,
		},
		{
			name:   "zero instance",
			given:  &edgev1alpha1.DeviceLink{},
			expect: true,
		},
	}

	for _, tc := range testCases {
		var ret = IsZero(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestIsActivating(t *testing.T) {
	var testCases = []struct {
		name   string
		given  metav1.Object
		expect bool
	}{
		{
			name: "existing instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expect: true,
		},
		{
			name:   "zero instance",
			given:  &edgev1alpha1.DeviceLink{},
			expect: false,
		},
		{
			name: "deleted instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "default",
					Name:              "test2",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = IsActivating(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
