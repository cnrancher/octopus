package suctioncup

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/metrics"
	"github.com/rancher/octopus/pkg/util/object"
)

func (m *manager) ExistAdaptor(name string) bool {
	return m.adaptors.Get(name) != nil
}

func (m *manager) Connect(by *edgev1alpha1.DeviceLink) (overwrite bool, berr error) {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return false, errors.New("adaptor name is empty")
	}

	// records metrics
	defer func() {
		if berr != nil {
			metrics.GetLimbMetricsRecorder().IncreaseConnectErrors(adaptorName)
		} else if !overwrite {
			metrics.GetLimbMetricsRecorder().IncreaseConnections(adaptorName)
		}
	}()

	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return false, errors.Errorf("could not find adaptor %s", adaptorName)
	}

	var name = object.GetNamespacedName(by)
	var ret, err = adaptor.CreateConnection(name)
	if err != nil {
		return false, errors.Wrapf(err, "failed to link %s", name)
	}
	return ret, nil
}

func (m *manager) Disconnect(by *edgev1alpha1.DeviceLink) (exist bool) {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return false
	}

	// records metrics
	defer func() {
		if exist {
			metrics.GetLimbMetricsRecorder().DecreaseConnections(adaptorName)
		}
	}()

	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return false
	}

	var name = object.GetNamespacedName(by)
	return adaptor.DeleteConnection(name)
}

func (m *manager) Send(data *unstructured.Unstructured, by *edgev1alpha1.DeviceLink) (berr error) {
	var adaptorName = by.Status.AdaptorName
	if adaptorName == "" {
		return errors.New("could not find blank name adaptor")
	}

	// records metrics
	var sendStartTS = time.Now()
	defer func() {
		metrics.GetLimbMetricsRecorder().ObserveSendLatency(adaptorName, time.Since(sendStartTS))
		if berr != nil {
			metrics.GetLimbMetricsRecorder().IncreaseSendErrors(adaptorName)
		}
	}()

	var adaptor = m.adaptors.Get(adaptorName)
	if adaptor == nil {
		return errors.Errorf("could not find adaptor %s", adaptorName)
	}

	var name = object.GetNamespacedName(by)
	var conn = adaptor.GetConnection(name)
	if conn == nil {
		return errors.Errorf("could not find connection %s", name)
	}

	// NB(thxCode) the data should never be nil
	var sendDevice, err = data.MarshalJSON()
	if err != nil {
		return errors.Wrapf(err, "could not marshal data as JSON")
	}
	var sendParameters []byte
	if by.Spec.Adaptor.Parameters != nil {
		sendParameters = by.Spec.Adaptor.Parameters.Raw
	}
	var sendModel = &by.Status.Model

	return conn.Send(sendParameters, sendModel, sendDevice)
}
