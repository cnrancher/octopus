package collection

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestStringSliceContain(t *testing.T) {
	var testCases = []struct {
		givenSlice []string
		given      string
		expect     bool
	}{
		{
			givenSlice: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
			given:  "Frank",
			expect: false,
		},
		{
			givenSlice: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
			given:  "Kobe",
			expect: true,
		},
	}

	for i, tc := range testCases {
		var ret = StringSliceContain(tc.givenSlice, tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestStringSliceRemove(t *testing.T) {
	var testCases = []struct {
		givenSlice []string
		given      string
		expect     []string
	}{
		{
			givenSlice: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
			given: "Frank",
			expect: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
		},
		{
			givenSlice: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
			given: "Kobe",
			expect: []string{
				"Jimmy",
				"Gucci",
				"Jack",
			},
		},
	}

	for i, tc := range testCases {
		var ret = StringSliceRemove(tc.givenSlice, tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
