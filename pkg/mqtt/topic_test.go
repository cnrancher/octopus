package mqtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/octopus/pkg/mqtt/api"
)

func TestSegmentTopic_RenderForPublishing(t *testing.T) {
	type given struct {
		topic     string
		operation api.MQTTMessageTopicOperation
		ref       corev1.ObjectReference
		appends   []map[string]string
	}

	var testOperation = api.MQTTMessageTopicOperation{
		Path: "path-xyz",
		Operator: &api.MQTTMessageTopicOperator{
			Read:  "status",
			Write: "set",
		},
	}
	var testRef = corev1.ObjectReference{
		Namespace: "default",
		Name:      "test",
		UID:       "uid-xyz",
	}

	var testCases = []struct {
		name     string
		given    given
		expected string
	}{
		{
			name: "render :namespace and :name",
			given: given{
				topic:     "cattle.io/octopus/:namespace/:name",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/default/test",
		},
		{
			name: "render :uid",
			given: given{
				topic:     "cattle.io/octopus/:uid",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/uid-xyz",
		},
		{
			name: "render :path",
			given: given{
				topic:     "cattle.io/octopus/:namespace/ccc/:name/:path",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/default/ccc/test/path-xyz",
		},
		{
			name: "render :operator",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/ccc/:name/:path",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/set/octopus/default/ccc/test/path-xyz",
		},
		{
			name: "render nothing",
			given: given{
				topic:     "cattle.io/octopus/static",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/static",
		},
		{
			name: "render redundant path",
			given: given{
				topic:     "cattle.io////octopus///static",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/static",
		},
		{
			name: "render overridden :path",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/:name/:path",
				operation: testOperation,
				ref:       testRef,
				appends: []map[string]string{
					{
						"path": "path-lmn",
					},
					{
						"path":     "path-abc",
						"operator": "_set_",
					},
				},
			},
			expected: "cattle.io/_set_/octopus/default/test/path-abc",
		},
		{
			name: "render overridden :path with blank string",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/:name/:path",
				operation: testOperation,
				ref:       testRef,
				appends: []map[string]string{
					{
						"path": "path-opq",
					},
					{
						"path":     "",
						"operator": "_set_",
					},
				},
			},
			expected: "cattle.io/_set_/octopus/default/test",
		},
	}

	for _, tc := range testCases {
		var st = NewSegmentTopic(tc.given.topic, tc.given.operation, tc.given.ref)
		var actual = st.RenderForPublish(tc.given.appends...)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestSegmentTopic_RenderForSubscribing(t *testing.T) {
	type given struct {
		topic     string
		operation api.MQTTMessageTopicOperation
		ref       corev1.ObjectReference
		appends   []map[string]string
	}

	var testOperation = api.MQTTMessageTopicOperation{
		Path: "path-xyz",
		Operator: &api.MQTTMessageTopicOperator{
			Read:  "status",
			Write: "set",
		},
	}
	var testRef = corev1.ObjectReference{
		Namespace: "default",
		Name:      "test",
		UID:       "uid-xyz",
	}

	var testCases = []struct {
		name     string
		given    given
		expected string
	}{
		{
			name: "render :namespace and :name",
			given: given{
				topic:     "cattle.io/octopus/:namespace/:name",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/default/test",
		},
		{
			name: "render :uid",
			given: given{
				topic:     "cattle.io/octopus/:uid",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/uid-xyz",
		},
		{
			name: "render :path",
			given: given{
				topic:     "cattle.io/octopus/:namespace/ccc/:name/:path",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/default/ccc/test/path-xyz",
		},
		{
			name: "render :operator",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/ccc/:name/:path",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/status/octopus/default/ccc/test/path-xyz",
		},
		{
			name: "render redundant path",
			given: given{
				topic:     "cattle.io////octopus///static",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/static",
		},
		{
			name: "render nothing",
			given: given{
				topic:     "cattle.io/octopus/static",
				operation: testOperation,
				ref:       testRef,
			},
			expected: "cattle.io/octopus/static",
		},
		{
			name: "render overridden :path",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/:name/:path",
				operation: testOperation,
				ref:       testRef,
				appends: []map[string]string{
					{
						"path": "path-lmn",
					},
					{
						"path":     "path-abc",
						"operator": "_status_",
					},
				},
			},
			expected: "cattle.io/_status_/octopus/default/test/path-abc",
		},
		{
			name: "render overridden :path with blank string",
			given: given{
				topic:     "cattle.io/:operator/octopus/:namespace/:name/:path",
				operation: testOperation,
				ref:       testRef,
				appends: []map[string]string{
					{
						"path": "path-opq",
					},
					{
						"path":     "",
						"operator": "_status_",
					},
				},
			},
			expected: "cattle.io/_status_/octopus/default/test",
		},
	}

	for _, tc := range testCases {
		var st = NewSegmentTopic(tc.given.topic, tc.given.operation, tc.given.ref)
		var actual = st.RenderForSubscribe(tc.given.appends...)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}
