package suctioncup

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/object"
)

func (m *manager) ExistAdaptor(name string) bool {
	return m.adaptors.Get(name) != nil
}

func (m *manager) Connect(by *edgev1alpha1.DeviceLink) error {
	var adaptor = m.adaptors.Get(by.Spec.Adaptor.Name)
	if adaptor == nil {
		return errors.Errorf("could not find adaptor %s", by.Spec.Adaptor.Name)
	}

	var name = object.GetNamespacedName(by)
	if err := adaptor.CreateConnection(name); err != nil {
		return errors.Wrapf(err, "failed to link %s", name)
	}
	return nil
}

func (m *manager) Disconnect(by *edgev1alpha1.DeviceLink) {
	var adaptor = m.adaptors.Get(by.Spec.Adaptor.Name)
	if adaptor == nil {
		return
	}

	var name = object.GetNamespacedName(by)
	adaptor.DeleteConnection(name)
}

func (m *manager) Send(data *unstructured.Unstructured, by *edgev1alpha1.DeviceLink) error {
	var adaptor = m.adaptors.Get(by.Spec.Adaptor.Name)
	if adaptor == nil {
		return errors.Errorf("could not find adaptor %s", by.Spec.Adaptor.Name)
	}

	var name = object.GetNamespacedName(by)
	var conn = adaptor.GetConnection(name)
	if conn == nil {
		return errors.Errorf("could not find connection %s", name)
	}

	var parametersBytes = by.Status.Adaptor.Parameters.DeepCopy().Raw
	var dataBytes, err = data.DeepCopy().MarshalJSON()
	if err != nil {
		return errors.Wrapf(err, "could not marshal data as JSON")
	}
	return conn.Send(parametersBytes, dataBytes)
}
