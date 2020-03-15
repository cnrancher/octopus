package physical

import (
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

type DataHandler func(name types.NamespacedName, status v1alpha1.ModbusDeviceStatus)
