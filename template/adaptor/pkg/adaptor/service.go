package adaptor

import (
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/template/adaptor/api/v1alpha1"
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
	// TODO implement the logic
	panic("implement me")
}
