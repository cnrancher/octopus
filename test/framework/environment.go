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
)

var testEmbeddedCluster EmbeddedCluster

func StartEnv(testEnv *envtest.Environment, writer io.Writer) (cfg *rest.Config, err error) {
	if !isUsingExistingCluster() {
		testEmbeddedCluster = NewEmbeddedCluster(KindCluster)
		if err := testEmbeddedCluster.Start(writer); err != nil {
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

func StopEnv(testEnv *envtest.Environment, writer io.Writer) error {
	if testEnv != nil {
		if err := testEnv.Stop(); err != nil {
			return err
		}
	}
	if !isUsingExistingCluster() {
		if testEmbeddedCluster != nil {
			if err := testEmbeddedCluster.Stop(writer); err != nil {
				return err
			}
		}
	}
	return nil
}

func isUsingExistingCluster() bool {
	return strings.EqualFold(os.Getenv(envUseExistingCluster), "true")
}
