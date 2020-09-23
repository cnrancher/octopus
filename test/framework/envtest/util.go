// +build test

package envtest

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/rancher/octopus/test/framework/envtest/cluster"
)

const (
	// Specify to test the cases on an existing cluster,
	// or create a local cluster for testing if blank.
	envUseExistingCluster = "USE_EXISTING_CLUSTER"

	// Specify the type of cluster.
	envClusterType = "CLUSTER_TYPE"
)

// GetConfig creates a *rest.Config for talking to a Kubernetes API server, the config precedence is as below:
// 1. KUBECONFIG environment variable pointing at files
// 2. $HOME/.kube/config if exists
// 3. In-cluster config if running in cluster
func GetConfig() (clientcmd.ClientConfig, error) {
	var loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		var u, err = user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not load config from current user as the current user is not found, %v", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
}

// IsUsingExistingCluster validates whether use an existing cluster for testing,
// if "USE_EXISTING_CLUSTER=true", the envtest will not create a local cluster for testing.
func IsUsingExistingCluster() bool {
	return strings.EqualFold(os.Getenv(envUseExistingCluster), "true")
}

// GetProvisionerType returns the provisioner type of testing cluster
// default is k3d.
func GetProvisionerType() cluster.ProvisionerType {
	var clusterType, exist = os.LookupEnv(envClusterType)
	if !exist {
		return cluster.ProvisionerTypeK3d
	}
	return cluster.ProvisionerType(clusterType)
}

// GetProvisioner returns the cluster provisioner from "CLUSTER_TYPE" environment,
// the default runtime is based on github.com/rancher/k3d.
func GetProvisioner() (cluster.Provisioner, error) {
	switch GetProvisionerType() {
	case cluster.ProvisionerTypeKind:
		return cluster.NewKind()
	default:
		return cluster.NewK3d()
	}
}
