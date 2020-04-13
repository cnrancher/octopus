package framework

import (
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	envUseExistingCluster = "USE_EXISTING_CLUSTER"
	envLocalClusterKind   = "LOCAL_CLUSTER_KIND"
)

var testLocalCluster LocalCluster

func StartEnv(rootDir string, testEnv *envtest.Environment, writer io.Writer) (cfg *rest.Config, err error) {
	if !IsUsingExistingCluster() {
		testLocalCluster = NewLocalCluster(GetLocalClusterKind())
		if err := testLocalCluster.Startup(rootDir, writer); err != nil {
			return nil, err
		}
	}
	if testEnv != nil {
		cfg, err = testEnv.Start()
		if err != nil {
			return nil, err
		}
		return
	}
	return nil, errors.New("test environment is nil")
}

func StopEnv(rootDir string, testEnv *envtest.Environment, writer io.Writer) error {
	if testEnv != nil {
		if err := testEnv.Stop(); err != nil {
			return err
		}
	}
	if !IsUsingExistingCluster() {
		if testLocalCluster != nil {
			if err := testLocalCluster.Cleanup(rootDir, writer); err != nil {
				return err
			}
		}
	}
	return nil
}

func IsUsingExistingCluster() bool {
	return strings.EqualFold(os.Getenv(envUseExistingCluster), "true")
}

func GetLocalClusterKind() ClusterKind {
	var kind = os.Getenv(envLocalClusterKind)
	if strings.EqualFold(kind, string(KindCluster)) {
		return KindCluster
	}
	return K3dCluster
}

func GetLocalCluster() LocalCluster {
	return testLocalCluster
}
