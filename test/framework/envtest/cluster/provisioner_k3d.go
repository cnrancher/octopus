// +build test

package cluster

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/k3d/v3/pkg/cluster"
	"github.com/rancher/k3d/v3/pkg/runtimes"
	"github.com/rancher/k3d/v3/pkg/types"

	"github.com/rancher/octopus/test/framework/finder"
)

// ProvisionerTypeK3d indicates the testing cluster is created by k3d.
const ProvisionerTypeK3d ProvisionerType = "k3d"

// NewK3d creates a github.com/rancher/k3d cluster.
func NewK3d() (*K3d, error) {
	var k3d = K3d{}
	if err := finder.Parse(&k3d); err != nil {
		return nil, errors.Wrap(err, "failed to parse K3d cluster argument from envs")
	}
	return &k3d, nil
}

type K3d struct {
	// Specify the image for running cluster,
	// configure in "K3S_IMAGE" env,
	// default is "rancher/k3s:v1.18.2-k3s1".
	Image string `env:"name=K3S_IMAGE,default=rancher/k3s:v1.18.2-k3s1"`

	// Specify the name of cluster,
	// configure in "K3S_CLUSTER_NAME" env,
	// default is "edge".
	ClusterName string `env:"name=K3S_CLUSTER_NAME,default=edge"`

	// Specify the amount of control-plane nodes,
	// configure in "K3S_CONTROL_PLANES" env,
	// default is "1".
	ControlPlanes int `env:"name=K3S_CONTROL_PLANES,default=1"`

	// Specify the amount of worker nodes,
	// configure in "K3S_WORKERS" env,
	// default is "2".
	Workers int `env:"name=K3S_WORKERS,default=2"`

	// Specify the wait timeout for bringing up cluster,
	// configure in "K3S_WAIT_TIMEOUT" env,
	// default is "5m".
	WaitTimeout time.Duration `env:"name=K3S_WAIT_TIMEOUT,default=5m"`

	// Specify the exported apiserver port for running cluster,
	// configure in "K3S_EXPORT_API_SERVER_PORT" env,
	// default is created randomly.
	ExportAPIServerPort int `env:"name=K3S_EXPORT_API_SERVER_PORT,fuzzPort"`

	// Specify the exported ingress http port for running cluster,
	// configure in "K3S_EXPORT_INGRESS_HTTP_PORT" env,
	// default is created randomly.
	ExportIngressHTTPPort int `env:"name=K3S_EXPORT_INGRESS_HTTP_PORT,fuzzPort"`

	// Specify the exported ingress https port for running cluster,
	// configure in "K3S_EXPORT_INGRESS_HTTPS_PORT" env,
	// default is created randomly.
	ExportIngressHTTPSPort int `env:"name=K3S_EXPORT_INGRESS_HTTPS_PORT,fuzzPort"`
}

func (c *K3d) String() string {
	return fmt.Sprintf("Name: %s, Image: %s, ControlPlanes: %d, Workers: %d", c.ClusterName, c.Image, c.ControlPlanes, c.Workers)
}

func (c *K3d) Startup() error {
	var ctx = context.TODO()
	var provider = runtimes.Docker

	// check if the cluster is existed
	var cls, err = getK3dCluster(ctx, provider, c.ClusterName)
	if err != nil {
		return err
	}

	// remove the existed cluster
	if cls != nil {
		err = cluster.ClusterDelete(ctx, provider, cls)
		if err != nil {
			return errors.Wrapf(err, "failed to clean the previous cluster")
		}
	}

	// create cluster configuration
	cls, err = renderK3sClusterConfiguration(c)
	if err != nil {
		return err
	}

	// create cluster
	err = cluster.ClusterCreate(ctx, provider, cls)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster")
	}

	// update kubeconfig
	_, err = cluster.KubeconfigGetWrite(ctx, provider, cls, "", &cluster.WriteKubeConfigOptions{UpdateExisting: true, OverwriteExisting: false, UpdateCurrentContext: true})
	if err != nil {
		return errors.Wrap(err, "failed to update kubeconfig file")
	}
	return nil
}

func (c *K3d) Cleanup() error {
	var ctx = context.TODO()
	var provider = runtimes.Docker

	// check if the cluster is existed
	var cls, err = getK3dCluster(ctx, provider, c.ClusterName)
	if err != nil {
		return err
	}

	if cls == nil {
		return nil
	}

	// delete cluster
	err = cluster.ClusterDelete(ctx, provider, cls)
	if err != nil {
		return errors.Wrap(err, "failed to clean")
	}

	// cleanup kubeconfig
	err = cluster.KubeconfigRemoveClusterFromDefaultConfig(ctx, cls)
	if err != nil {
		return errors.Wrap(err, "failed to cleanup kubeconfig file")
	}
	return nil
}

func (c *K3d) AddWorker(name string) error {
	var ctx = context.TODO()
	var provider = runtimes.Docker

	// check if the cluster is existed
	var cls, err = getK3dCluster(ctx, provider, c.ClusterName)
	if err != nil {
		return err
	}

	if cls == nil {
		return nil
	}

	var count, _ = cls.AgentCountRunning()

	// create node
	var node = &types.Node{
		Name:  fmt.Sprintf("%s-%s-agent-%d", types.DefaultObjectNamePrefix, c.ClusterName, count),
		Role:  types.AgentRole,
		Image: c.Image,
		Args:  []string{fmt.Sprintf("--node-name=%s", name)},
	}

	// add node to cluster
	err = cluster.NodeAddToCluster(ctx, provider, node, cls, types.NodeCreateOpts{Timeout: c.WaitTimeout, Wait: true})
	if err != nil {
		return errors.Wrap(err, "failed to add new node ")
	}

	return nil
}

func (c *K3d) IsLocalCluster() bool {
	return true
}

func renderK3sClusterConfiguration(config *K3d) (*types.Cluster, error) {
	var cls = &types.Cluster{
		Name: config.ClusterName,
		CreateClusterOpts: &types.ClusterCreateOpts{
			Timeout:       config.WaitTimeout,
			WaitForServer: true,
		},
		ServerLoadBalancer: &types.Node{
			Role: types.LoadBalancerRole,
		},
		ExposeAPI: types.ExposeAPI{
			HostIP: types.DefaultAPIHost,
			Host:   types.DefaultAPIHost,
			Port:   strconv.Itoa(config.ExportAPIServerPort),
		},
	}

	for i := 0; i < config.ControlPlanes; i++ {
		var node = &types.Node{
			Role:  types.ServerRole,
			Image: config.Image,
		}
		if i == 0 {
			node.Args = []string{fmt.Sprintf("--node-name=%s-control-plane", config.ClusterName)}
			node.Ports = []string{fmt.Sprintf("%d:80", config.ExportIngressHTTPPort), fmt.Sprintf("%d:443", config.ExportIngressHTTPSPort)}
		} else {
			node.Args = []string{fmt.Sprintf("--node-name=%s-control-plane%d", config.ClusterName, i)}
		}
		cls.Nodes = append(cls.Nodes, node)
	}

	for i := 0; i < config.Workers; i++ {
		var node = &types.Node{
			Role:  types.AgentRole,
			Image: config.Image,
		}
		if i == 0 {
			node.Args = []string{fmt.Sprintf("--node-name=%s-worker", config.ClusterName)}
		} else {
			node.Args = []string{fmt.Sprintf("--node-name=%s-worker%d", config.ClusterName, i)}
		}
		cls.Nodes = append(cls.Nodes, node)
	}

	return cls, nil
}

func getK3dCluster(ctx context.Context, runtime runtimes.Runtime, clusterName string) (*types.Cluster, error) {
	var cls, err = cluster.ClusterGet(ctx, runtime, &types.Cluster{Name: clusterName})
	if err != nil {
		if err.Error() == fmt.Sprintf("No nodes found for cluster '%s'", clusterName) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to get %s cluster", clusterName)
	}
	return cls, nil
}
