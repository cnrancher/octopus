package fieldpath

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestExtractDeviceLinkFieldPathAsBytes(t *testing.T) {
	type given struct {
		link      *edgev1alpha1.DeviceLink
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
			name: "metadata.annotations",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.annotations",
			},
			expected: expected{
				ret: []byte(`annotation-key-1="v1";annotation-key-2="v2"`),
			},
		},
		{
			name: "metadata.name",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.name",
			},
			expected: expected{
				ret: []byte(`test`),
			},
		},
		{
			name: "metadata.name['lb3']",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.labels['lb3']",
			},
			expected: expected{
				ret: nil,
			},
		},
		{
			name: "status.nodeName",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "status.nodeName",
			},
			expected: expected{
				err: errors.Errorf("unsupported fieldPath: status.nodeName"),
			},
		},
		{
			name: "status.nodeHostName",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "status.nodeHostName",
			},
			expected: expected{
				ret: []byte(`test-node-1`),
			},
		},
		{
			name: "status.xxxx",
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: `status.xxxx`,
			},
			expected: expected{
				err: errors.Errorf("unsupported fieldPath: status.xxxx"),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = ExtractDeviceLinkFieldPathAsBytes(tc.given.link, tc.given.fieldPath)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}
