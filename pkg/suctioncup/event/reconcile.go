package event

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type Response struct {
	Requeue      bool
	RequeueAfter time.Duration
}

type RequestAdaptorStatus struct {
	Name       string
	Registered bool
}

type RequestConnectionStatus struct {
	AdaptorName string
	Name        types.NamespacedName
	Data        []byte
	Error       error
	Closed      bool
}

type adaptorRegistered struct {
	name string
}

type adaptorUnregistered struct {
	name string
}

type connectionReceivedData struct {
	adaptorName string
	name        types.NamespacedName
}

type connectionReceivedError struct {
	adaptorName string
	name        types.NamespacedName
}

type connectionClosed struct {
	adaptorName string
	name        types.NamespacedName
}
