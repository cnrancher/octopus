package collection

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestStringMapCopy(t *testing.T) {
	var testCases = []struct {
		given  map[string]string
		expect map[string]string
	}{
		{
			given: map[string]string{
				"k1": "v1",
			},
			expect: map[string]string{
				"k1": "v1",
			},
		},
		{
			given: map[string]string{
				"k2": "v2",
				"k3": "v3",
			},
			expect: map[string]string{
				"k2": "v2",
				"k3": "v3",
			},
		},
	}

	for i, tc := range testCases {
		var ret = StringMapCopy(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}

		ret["xyz"] = "zyx"
		if reflect.DeepEqual(ret, tc.given) {
			t.Errorf("case %v: expected is pointing to the same map with given", i+1)
		}
	}
}

func TestStringMapCopyInto(t *testing.T) {
	type given struct {
		source map[string]string
		target map[string]string
	}
	var testCases = []struct {
		given  given
		expect map[string]string
	}{
		{
			given: given{
				source: map[string]string{
					"s1": "s1",
				},
				target: map[string]string{
					"d1": "d1",
				},
			},
			expect: map[string]string{
				"s1": "s1",
				"d1": "d1",
			},
		},
		{
			given: given{
				source: map[string]string{
					"s2": "s2",
				},
				target: map[string]string{
					"d2": "d2",
				},
			},
			expect: map[string]string{
				"s2": "s2",
				"d2": "d2",
			},
		},
	}

	for i, tc := range testCases {
		var ret = StringMapCopyInto(tc.given.source, tc.given.target)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}

		ret["xyz"] = "zyx"
		if reflect.DeepEqual(ret, tc.given.source) {
			t.Errorf("case %v: expected is pointing to the same map with given source", i+1)
		}
		if !reflect.DeepEqual(ret, tc.given.target) {
			t.Errorf("case %v: expected is not pointing to the same map with given target", i+1)
		}
	}
}

func TestDiffStringMap(t *testing.T) {
	type given struct {
		left  map[string]string
		right map[string]string
	}

	var testCases = []struct {
		given  given
		expect bool
	}{
		{
			given: given{
				left: map[string]string{
					"a": "b",
					"c": "d",
				},
				right: map[string]string{
					"c": "d",
					"a": "b",
				},
			},
			expect: false,
		},
		{
			given: given{
				left: map[string]string{
					"a": "b",
					"c": "d",
				},
				right: map[string]string{
					"c": "d",
				},
			},
			expect: true,
		},
	}

	for i, tc := range testCases {
		var ret = DiffStringMap(tc.given.left, tc.given.right)
		if ret != tc.expect {
			t.Errorf("case %v: expected %v, got %v", i+1, tc.expect, ret)
		}
	}
}

func TestFormatStringMap(t *testing.T) {
	type given struct {
		m        map[string]string
		splitter string
	}
	var testCases = []struct {
		given  given
		expect string
	}{
		{
			given: given{
				m: map[string]string{
					"c": "d",
					"x": "z",
					"a": "b",
				},
				splitter: "",
			},
			expect: `a="b",c="d",x="z"`,
		},
		{
			given: given{
				m: map[string]string{
					"c": "d",
					"x": "z",
					"a": "b",
				},
				splitter: ";",
			},
			expect: `a="b";c="d";x="z"`,
		},
	}

	for i, tc := range testCases {
		var ret = FormatStringMap(tc.given.m, tc.given.splitter)
		if ret != tc.expect {
			t.Errorf("case %v: expected %v, got %v", i+1, tc.expect, ret)
		}
	}
}
