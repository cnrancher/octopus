package model

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetCRDNameOfGroupVersionKind(t *testing.T) {
	var testCases = []struct {
		given  schema.GroupVersionKind
		expect string
	}{
		{
			given: schema.GroupVersionKind{
				Group:   "test.io",
				Kind:    "K1",
				Version: "v1",
			},
			expect: "k1s.test.io",
		},
		{
			given: schema.GroupVersionKind{
				Group:   "test.io",
				Kind:    "K2",
				Version: "v1",
			},
			expect: "k2s.test.io",
		},
	}

	for i, tc := range testCases {
		var ret = GetCRDNameOfGroupVersionKind(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %d: expected %s, got %s", i, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
