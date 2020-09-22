package adaptor

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/mqtt/pkg/physical"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/adaptor/log"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/object"
)

func NewService() *Service {
	mqtt.SetLogger(log.GetLogger())

	var scheme = k8sruntime.NewScheme()
	// register v1alpha1 scheme into runtime scheme.
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
	var holder physical.Device
	defer func() {
		if holder != nil {
			holder.Shutdown()
		}
	}()

	for {
		var req, err = server.Recv()
		if err != nil {
			if !connection.IsClosed(err) {
				log.Error(err, "Failed to receive connect request from Limb")
				return status.Error(codes.Unknown, "shutdown connection as receiving error from Limb")
			}
			return nil
		}

		// validates model GVK
		var model = req.GetModel()
		if model == nil {
			return status.Error(codes.InvalidArgument, "invalid empty model")
		}
		var modelGVK = model.GroupVersionKind()
		if modelGVK.Group != "devices.edge.cattle.io" {
			return status.Errorf(codes.InvalidArgument, "invalid model group: %s", modelGVK.Group)
		}

		// processes device
		switch modelGVK.Kind {
		case "MQTTDevice":
			// gets device spec
			var device v1alpha1.MQTTDevice
			if err := converter.UnmarshalJSON(req.GetDevice(), &device); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to unmarshal device: %v", err)
			}

			// creates device handler
			if holder == nil {
				// gets device namespaced name
				var deviceName = object.GetNamespacedName(&device)
				if deviceName.Namespace == "" || deviceName.Name == "" {
					return status.Error(codes.InvalidArgument, "failed to recognize the empty device as the namespace/name is blank")
				}

				// gets log
				var logger = log.WithValues("mqtt device", deviceName)

				// creates handler for syncing to limb
				var toLimb = func(in *v1alpha1.MQTTDevice, internalError error) error {
					var resp *api.ConnectResponse
					if internalError != nil {
						// feedback error message
						resp = &api.ConnectResponse{ErrorMessage: internalError.Error()}
					} else {
						// send device by {name, namespace, status} tuple
						var device = &v1alpha1.MQTTDevice{}
						device.Namespace = in.Namespace
						device.Name = in.Name
						device.Status = in.Status
						// convert device to json bytes
						var deviceBytes = s.toJSON(device)
						resp = &api.ConnectResponse{Device: deviceBytes}
					}
					// send device to limb
					if err := server.Send(resp); err != nil {
						return status.Errorf(codes.Unknown, "failed to send device to limb, %v", err)
					}
					return nil
				}

				holder = physical.NewDevice(logger, device.ObjectMeta, toLimb)
			}

			// configures device
			if err := holder.Configure(req.GetReferencesHandler(), &device); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to configure the device: %v", err)
			}
		default:
			return status.Errorf(codes.InvalidArgument, "invalid model kind: %s", modelGVK.Kind)
		}
	}
}
