package physical

import (
	"sync"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
)

type SubscriptionInfo struct {
	Topic       string
	Payload     []byte
	PayloadType v1alpha1.PayloadType
}

type SubscriptionMap struct {
	sync.Map
}
