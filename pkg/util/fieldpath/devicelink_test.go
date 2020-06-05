package fieldpath

import (
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
	type expect struct {
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
		given  given
		expect expect
	}{
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.annotations",
			},
			expect: expect{
				ret: []byte(`annotation-key-1="v1";annotation-key-2="v2"`),
				err: nil,
			},
		},
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.name",
			},
			expect: expect{
				ret: []byte(`test`),
				err: nil,
			},
		},
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: "metadata.labels['lb3']",
			},
			expect: expect{
				ret: nil,
				err: nil,
			},
		},
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: `status.nodeName`,
			},
			expect: expect{
				ret: nil,
				err: errors.Errorf("unsupported fieldPath: status.nodeName"),
			},
		},
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: `status.nodeHostName`,
			},
			expect: expect{
				ret: []byte(`test-node-1`),
				err: nil,
			},
		},
		{
			given: given{
				link:      targetObject.DeepCopy(),
				fieldPath: `status.xxxx`,
			},
			expect: expect{
				ret: nil,
				err: errors.Errorf("unsupported fieldPath: status.xxxx"),
			},
		},
	}

	for i, tc := range testCases {
		var actualBytes, actualErr = ExtractDeviceLinkFieldPathAsBytes(tc.given.link, tc.given.fieldPath)
		if actualErr != nil {
			if tc.expect.err != nil {
				assert.EqualError(t, actualErr, tc.expect.err.Error(), "case %v", i+1)
			} else {
				assert.NoError(t, actualErr, "case %v ", i+1)
			}
		}
		assert.Equal(t, actualBytes, tc.expect.ret, "case %v", i+1)
	}
}
