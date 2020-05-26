package adaptor

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rancher/octopus/adaptors/agent-device/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/agent-device/pkg/physical"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/connection"
	"github.com/rancher/octopus/pkg/util/object"
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = logr.NewDelegatingLogger(nil)

func init() {
	log.Fulfill(zap.New(
		zap.UseDevMode(true),
		zap.Level(func() *uberzap.AtomicLevel {
			level := uberzap.NewAtomicLevelAt(uberzapcore.DebugLevel)
			return &level
		}()),
	))
}

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

		// validate parameters
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
		var agent v1alpha1.AgentDeviceGroup
		if err := jsoniter.Unmarshal(req.GetDevice(), &agent); err != nil {
			return status.Errorf(codes.InvalidArgument, "failed to unmarshal device: %v", err)
		}

		// process device
		if device == nil {
			var deviceName = object.GetNamespacedName(&agent)
			var dataHandler = func(name types.NamespacedName, status v1alpha1.AgentDeviceGroupStatus) {
				// send device by {name, namespace, status} tuple
				var resp v1alpha1.AgentDeviceGroup
				resp.Namespace = name.Namespace
				resp.Name = name.Name
				resp.Status = status

				// convert device to json bytes
				var respBytes = s.toJSON(&resp)

				// send device
				if err := server.Send(&api.ConnectResponse{Device: respBytes}); err != nil {
					if !connection.IsClosed(err) {
						log.Error(err, "Failed to send response to connection")
					}
				}
			}

			cfg := ctrl.GetConfigOrDie()
			clientSet, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				log.Error(err, "Unable to get clientSet")
				return status.Errorf(codes.FailedPrecondition, "failed to get clientSet: %v", err)
			}

			device = physical.NewDevice(
				log.WithValues("device", deviceName),
				deviceName,
				clientSet,
				dataHandler)

			device.On()
		}
		device.Configure(agent.Spec)
	}
}
