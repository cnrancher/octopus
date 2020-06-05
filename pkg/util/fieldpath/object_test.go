package fieldpath

import (
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
				obj:       targetObject.DeepCopy(),
				fieldPath: "metadata.labels",
			},
			expect: expect{
				ret: []byte(`lb1="v1";lb2="v2"`),
				err: nil,
			},
		},
		{
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: `metadata.annotations['annotation-key-1']`,
			},
			expect: expect{
				ret: []byte(`v1`),
				err: nil,
			},
		},
		{
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: `metadata.namespace`,
			},
			expect: expect{
				ret: []byte(`default`),
				err: nil,
			},
		},
		{
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: `metadata.annotations['annotation-key-3']`,
			},
			expect: expect{
				ret: nil,
				err: nil,
			},
		},
		{
			given: given{
				obj:       targetObject.DeepCopy(),
				fieldPath: `status.nodeName`,
			},
			expect: expect{
				ret: nil,
				err: errors.Errorf("unsupported fieldPath: status.nodeName"),
			},
		},
	}

	for i, tc := range testCases {
		var actualBytes, actualErr = ExtractObjectFieldPathAsBytes(tc.given.obj, tc.given.fieldPath)
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
