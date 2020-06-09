package adaptor

import (
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/dummy/pkg/physical"
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

		// validate model GVK
		var model = req.GetModel()
		if model == nil {
			return status.Error(codes.InvalidArgument, "invalid empty model")
		}
		var modelGVK = model.GroupVersionKind()
		if modelGVK.Group != v1alpha1.GroupVersion.Group {
			return status.Errorf(codes.InvalidArgument, "invalid model group: %s", modelGVK.Group)
		}
		// NB(thxCode) the version of model can use to make compatible, the kind of model can use to determine the type of object.

		// process device
		switch modelGVK.Kind {
		case "DummySpecialDevice":
			// get device spec
			var device v1alpha1.DummySpecialDevice
			if err := jsoniter.Unmarshal(req.GetDevice(), &device); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to unmarshal device: %v", err)
			}

			// create device handler
			if holder == nil {
				// get device NamespacedName
				var deviceName = object.GetNamespacedName(&device)
				if deviceName.Namespace == "" || deviceName.Name == "" {
					return status.Error(codes.InvalidArgument, "failed to recognize the empty device as the namespace/name is blank")
				}

				// create handler for sync to limb
				var toLimb = func(in *v1alpha1.DummySpecialDevice) {
					// convert device to json bytes
					var respBytes = s.toJSON(in)

					// send device to limb
					if err := server.Send(&api.ConnectResponse{Device: respBytes}); err != nil {
						if !connection.IsClosed(err) {
							log.Error(err, "Failed to send response to connection", "device", deviceName)
						}
					}
				}

				holder = physical.NewSpecialDevice(
					log.WithValues("device", deviceName),
					&device,
					toLimb,
				)
			}

			// configure device
			if err := holder.Configure(req.GetReferencesHandler(), device.Spec); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to configure the device: %v", err)
			}

		case "DummyProtocolDevice":
			// get device spec
			var device v1alpha1.DummyProtocolDevice
			if err := jsoniter.Unmarshal(req.GetDevice(), &device); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to unmarshal device: %v", err)
			}

			// create device handler
			if holder == nil {
				// get device NamespacedName
				var deviceName = object.GetNamespacedName(&device)
				if deviceName.Namespace == "" || deviceName.Name == "" {
					return status.Error(codes.InvalidArgument, "failed to recognize the empty device as the namespace/name is blank")
				}

				// create handler for sync to limb
				var toLimb = func(in *v1alpha1.DummyProtocolDevice) {
					// convert device to json bytes
					var respBytes = s.toJSON(in)

					// send device to limb
					if err := server.Send(&api.ConnectResponse{Device: respBytes}); err != nil {
						if !connection.IsClosed(err) {
							log.Error(err, "Failed to send response to connection", "device", deviceName)
						}
					}
				}

				holder = physical.NewProtocolDevice(
					log.WithValues("device", deviceName),
					&device,
					toLimb,
				)
			}

			// configure device
			if err := holder.Configure(req.GetReferencesHandler(), device.Spec); err != nil {
				return status.Errorf(codes.InvalidArgument, "failed to configure the device: %v", err)
			}

		default:
			return status.Errorf(codes.InvalidArgument, "invalid model kind: %s", modelGVK.Kind)
		}
	}
}
