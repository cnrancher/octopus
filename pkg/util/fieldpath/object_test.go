package fieldpath

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestExtractObjectFieldPathAsBytes(t *testing.T) {
	type given struct {
		obj       runtime.Object
		fieldPath string
	}
	type expected struct {
		ret []byte
		err error
	}

	var targetObject = &edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			UID:       "uid1",
			Annotations: map[string]string{
				"annotation-key-1": "v1",
				"annotation-key-2": "v2",
			},
			Labels: map[string]string{
				"lb1": "v1",
				"lb2": "v2",
			},
		},
		Status: edgev1alpha1.DeviceLinkStatus{
			NodeName:       "edge-worker",
			NodeHostName:   "test-node-1",
			NodeInternalIP: "192.168.1.34",
		},
	}

	var testCases = []struct {
		name     string
		given    given
		expected expected
	}{
		{
			name: "metadata.labels",
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: "metadata.labels",
			},
			expected: expected{
				ret: []byte(`lb1="v1";lb2="v2"`),
			},
		},
		{
			name: "metadata.annotations['annotation-key-1']",
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: "metadata.annotations['annotation-key-1']",
			},
			expected: expected{
				ret: []byte(`v1`),
			},
		},
		{
			name: "metadata.namespace",
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: "metadata.namespace",
			},
			expected: expected{
				ret: []byte(`default`),
			},
		},
		{
			name: "metadata.annotations['annotation-key-3']",
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: "metadata.annotations['annotation-key-3']",
			},
			expected: expected{
				ret: nil,
			},
		},
		{
			name: "status.nodeName",
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: "status.nodeName",
			},
			expected: expected{
				err: errors.Errorf("unsupported fieldPath: status.nodeName"),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = ExtractObjectFieldPathAsBytes(tc.given.obj, tc.given.fieldPath)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}
