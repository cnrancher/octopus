package suctioncup

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/suctioncup/event"
)

type Manager interface {
	// Start starts the suction cup manager.
	Start(<-chan struct{}) error

	// RegisterAdaptorHandler registers a handler for reconciling the adaptor events.
	RegisterAdaptorHandler(handler event.AdaptorHandler)

	// RegisterAdaptorHandler registers a handler for reconciling the connection events.
	RegisterConnectionHandler(handler event.ConnectionHandler)

	// GetNeurons returns the Neurons.
	GetNeurons() Neurons
}

type Neurons interface {
	// ExistAdaptor judges whether the adaptor of target exist.
	ExistAdaptor(name string) bool

	// Connect starts a connection by link, the return "overwrite" represents whether to overwrite an existing connection.
	Connect(referencesData map[string]map[string][]byte, device *unstructured.Unstructured, by *edgev1alpha1.DeviceLink) error

	// Disconnect stops a connection by link
	Disconnect(by *edgev1alpha1.DeviceLink)
}
