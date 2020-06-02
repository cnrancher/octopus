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

func (m *manager) Connect(by *edgev1alpha1.DeviceLink) (bool, error) {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return false, errors.New("adaptor name is empty")
	}
	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return false, errors.Errorf("could not find adaptor %s", adaptorName)
	}

	var name = object.GetNamespacedName(by)
	var overwrite, err = adaptor.CreateConnection(name)
	if err != nil {
		return false, errors.Wrapf(err, "failed to link %s", name)
	}
	return overwrite, nil
}

func (m *manager) Disconnect(by *edgev1alpha1.DeviceLink) bool {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return false
	}
	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return false
	}

	var name = object.GetNamespacedName(by)
	return adaptor.DeleteConnection(name)
}

func (m *manager) Send(data *unstructured.Unstructured, by *edgev1alpha1.DeviceLink) error {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return errors.New("could not find blank name adaptor")
	}
	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return errors.Errorf("could not find adaptor %s", adaptorName)
	}

	var name = object.GetNamespacedName(by)
	var conn = adaptor.GetConnection(name)
	if conn == nil {
		return errors.Errorf("could not find connection %s", name)
	}
	var parametersBytes []byte
	if by.Status.Adaptor.Parameters != nil {
		parametersBytes = by.Status.Adaptor.Parameters.DeepCopy().Raw
	}
	var dataBytes, err = data.DeepCopy().MarshalJSON()
	if err != nil {
		return errors.Wrapf(err, "could not marshal data as JSON")
	}

	return conn.Send(parametersBytes, &by.Status.Model, dataBytes)
}
