// +build test

package predicate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type runtimeMetaObject interface {
	runtime.Object
	metav1.Object
}

func generateUpdateEvent(o, n runtimeMetaObject) event.UpdateEvent {
	return event.UpdateEvent{
		MetaOld:   o,
		ObjectOld: o,
		MetaNew:   n,
		ObjectNew: n,
	}
}
