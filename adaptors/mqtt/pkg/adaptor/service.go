package adaptor

import (
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/mqtt/pkg/physical"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/util/object"
)

func NewService() *Service {
	var scheme = k8sruntime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	return &Service{
		scheme: scheme,
	}
}

type Service struct {
	scheme *k8sruntime.Scheme
}

func (s *Service) toJSON(in metav1.Object) []byte {
	var out = unstructured.Unstructured{Object: make(map[string]interface{})}
	// NB(thxCode) scheme conversion can keep the typemeta of an object,
	// provided that the object type has been registered in scheme first.
	_ = s.scheme.Convert(in, &out, nil)
	var bytes, _ = out.MarshalJSON()
	return bytes
}

func (s *Service) Connect(server api.Connection_ConnectServer) error {
	var device physical.Device
	defer func() {
		if device != nil {
			device.Shutdown()
		}
	}()

	for {

		var req, err = server.Recv()
		if err != nil {
			if !connection.IsClosed(err) {
				log.Error(err, "Failed to receive connect request from Limb")
				return status.Errorf(codes.Unknown, "shutdown connection as receiving error from Limb")
			}
			return nil
		}

		var parameters = physical.DefaultParameters()
		if req.GetParameters() != nil {
			if err := jsoniter.Unmarshal(req.GetParameters(), &parameters); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to unmarshal parameters: %v", err)
			}
		}

		if err := parameters.Validate(); err != nil {
			return status.Errorf(codes.InvalidArgument, "failed to validate parameters: %v", err)
		}

		// validate device
		var mqtt v1alpha1.MqttDevice
		if err := jsoniter.Unmarshal(req.GetDevice(), &mqtt); err != nil {
			return status.Errorf(codes.InvalidArgument, "failed to unmarshal device: %v", err)
		}

		if device == nil {
			var deviceName = object.GetNamespacedName(&mqtt)
			if deviceName.Namespace == "" || deviceName.Name == "" {
				return status.Error(codes.InvalidArgument, "failed to recognize the empty device as the namespace/name is blank")
			}

			var dataHandler = func(name types.NamespacedName, status v1alpha1.MqttDeviceStatus) {
				// send device by {name, namespace, status} tuple
				var resp v1alpha1.MqttDevice
				resp.Namespace = name.Namespace
				resp.Name = name.Name
				resp.Status = status

				// convert device to json bytes
				var respBytes = s.toJSON(&resp)

				log.Info("dataHandler device update", "MqttDevice", string(respBytes))

				// send device
				if err := server.Send(&api.ConnectResponse{Device: respBytes}); err != nil {
					if !connection.IsClosed(err) {
						log.Error(err, "Failed to send response to connection")
					}
				}
			}

			mqttClient, err := physical.NewMqttClient(mqtt.Name, mqtt.Spec.Config)
			if err != nil {
				log.Error(err, "connect receive new device NewMqttClient error")
				return status.Errorf(codes.InvalidArgument, "failed to connect mqtt: %v", err)
			}

			device = physical.NewDevice(
				log.WithValues("device", deviceName),
				&mqtt,
				dataHandler,
				mqttClient,
			)

			go device.On()

			log.Info("connect receive new device success", "name", mqtt.Name)

		} else {
			device.Configure(&mqtt.Spec)
		}

	}
}
