package mqtt

import (
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/octopus/pkg/mqtt/api"
	"github.com/rancher/octopus/pkg/util/collection"
)

type SegmentTopic interface {
	// RenderForSubscribe renders the topic with templates started with low priority.
	// and returns the result for subscribing.
	RenderForSubscribe(renders ...map[string]string) string

	// RenderForPublish renders the topic with templates started with low priority,
	// and returns the result for publishing.
	RenderForPublish(renders ...map[string]string) string
}

type segmentTopic struct {
	segments  []string
	operation api.MQTTMessageTopicOperation
}

func (t segmentTopic) render(root map[string]string, appends ...map[string]string) string {
	for _, r := range appends {
		collection.StringMapCopyInto(r, root)
	}

	var segments = make([]string, 0, len(t.segments))
	for _, seg := range t.segments {
		if seg != ":" {
			if seg[0] == ':' {
				var r, exist = root[seg[1:]]
				if !exist {
					continue
				}
				seg = r
			}
			if seg != "" {
				segments = append(segments, seg)
			}
		}
	}
	return path.Join(segments...)
}

func (t segmentTopic) RenderForSubscribe(renders ...map[string]string) string {
	var globalRender = make(map[string]string, 2)
	globalRender["path"] = t.operation.Path
	if t.operation.Operator != nil {
		var read = t.operation.Operator.Read
		if read == "null" {
			read = ""
		}
		globalRender["operator"] = read
	}
	return t.render(globalRender, renders...)
}

func (t segmentTopic) RenderForPublish(renders ...map[string]string) string {
	var globalRender = make(map[string]string, 2)
	globalRender["path"] = t.operation.Path
	if t.operation.Operator != nil {
		var write = t.operation.Operator.Write
		if write == "null" {
			write = ""
		}
		globalRender["operator"] = write
	}
	return t.render(globalRender, renders...)
}

func NewSegmentTopic(topic string, operation api.MQTTMessageTopicOperation, ref corev1.ObjectReference) SegmentTopic {
	var segments = strings.Split(topic, "/")
	var newSegments = make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if seg != ":" {
			if seg[0] == ':' {
				switch seg[1:] {
				case "namespace":
					seg = ref.Namespace
				case "name":
					seg = ref.Name
				case "uid":
					seg = string(ref.UID)
				}
			}
			if seg != "" {
				newSegments = append(newSegments, seg)
			}
		}
	}
	return segmentTopic{
		segments:  newSegments,
		operation: operation,
	}
}
