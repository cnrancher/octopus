package handler

import (
	"runtime"

	"github.com/go-logr/logr"
	util "k8s.io/apimachinery/pkg/util/runtime"
)

func NewPanicsLogHandler(log logr.Logger) func(interface{}) {
	return func(r interface{}) {
		const size = 64 << 10
		stacktrace := make([]byte, size)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
		if _, ok := r.(string); ok {
			log.Error(nil, "observed a panic: %s\n%s", r, stacktrace)
		} else {
			log.Error(nil, "Observed a panic: %#v (%v)\n%s", r, r, stacktrace)
		}
	}
}

func init() {
	util.PanicHandlers = []func(interface{}){}
}
