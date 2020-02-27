package model

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewInstanceOfTypeMeta(t *testing.T) {
	var testCases = []struct {
		given  metav1.TypeMeta
		expect unstructured.Unstructured
	}{
		{
			given: metav1.TypeMeta{
				Kind:       "K1",
				APIVersion: "test.io/v1",
			},
			expect: unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K1",
					"apiVersion": "test.io/v1",
				},
			},
		},
		{
			given: metav1.TypeMeta{
				Kind:       "K2",
				APIVersion: "test.io/v1alpha1",
			},
			expect: unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K2",
					"apiVersion": "test.io/v1alpha1",
				},
			},
		},
	}

	for i, tc := range testCases {
		var ret = NewInstanceOfTypeMeta(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %d: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestNewInstanceOfType(t *testing.T) {
	var testCases = []struct {
		given  metav1.Type
		expect unstructured.Unstructured
	}{
		{
			given: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K1",
					"apiVersion": "test.io/v1",
				},
			},
			expect: unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K1",
					"apiVersion": "test.io/v1",
				},
			},
		},
		{
			given: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K2",
					"apiVersion": "test.io/v1alpha1",
				},
			},
			expect: unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K2",
					"apiVersion": "test.io/v1alpha1",
				},
			},
		},
	}

	for i, tc := range testCases {
		var ret = NewInstanceOfType(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %d: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}
