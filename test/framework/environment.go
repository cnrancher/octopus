// +build test

package framework

import (
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/rancher/octopus/test/framework/cluster"
)

const (
	envUseExistingCluster = "USE_EXISTING_CLUSTER"
	envClusterType        = "CLUSTER_TYPE"
)

func StartEnv(rootDir string, testEnv *envtest.Environment, writer io.Writer) (cfg *rest.Config, err error) {
	if !IsUsingExistingCluster() {
		if err := GetCluster().Startup(rootDir, writer); err != nil {
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
		if err := GetCluster().Cleanup(rootDir, writer); err != nil {
			return err
		}
	}
	return nil
}

func IsUsingExistingCluster() bool {
	return strings.EqualFold(os.Getenv(envUseExistingCluster), "true")
}

func GetCluster() cluster.Cluster {
	var cls = cluster.Cluster(os.Getenv(envClusterType))
	if cls == "" {
		return cluster.K3d
	}
	return cls
}
