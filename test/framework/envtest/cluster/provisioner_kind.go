// +build test

package cluster

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"

	"github.com/rancher/octopus/test/framework/envtest/cluster/internal/kind/logs"
	"github.com/rancher/octopus/test/framework/finder"
)

// ProvisionerTypeKind indicates the testing cluster is created by kind.
const ProvisionerTypeKind ProvisionerType = "kind"

// NewKind creates a kubernetes-sigs/kind cluster.
func NewKind() (*Kind, error) {
	var kind = Kind{}
	if err := finder.Parse(&kind); err != nil {
		return nil, errors.Wrap(err, "failed to parse Kind cluster argument from envs")
	}
	return &kind, nil
}

type Kind struct {
	// Specify the image for running cluster,
	// configure in "KIND_IMAGE" env,
	// default is "kindest/node:v1.18.2".
	Image string `env:"name=KIND_IMAGE,default=kindest/node:v1.18.2"`

	// Specify the name of cluster,
	// configure in "KIND_CLUSTER_NAME" env,
	// default is "edge".
	ClusterName string `env:"name=KIND_CLUSTER_NAME,default=edge"`

	// Specify the amount of control-plane nodes,
	// configure in "KIND_CONTROL_PLANES" env,
	// default is "1".
	ControlPlanes int `env:"name=KIND_CONTROL_PLANES,default=1"`

	// Specify the amount of worker nodes,
	// configure in "KIND_WORKERS" env,
	// default is "2".
	Workers int `env:"name=KIND_WORKERS,default=3"`

	// Specify the wait timeout for bringing up cluster,
	// configure in "KIND_WAIT_TIMEOUT" env,
	// default is "5m".
	WaitTimeout time.Duration `env:"name=KIND_WAIT_TIMEOUT,default=5m"`

	// Specify the path of preset cluster configuration,
	// configure in "KIND_CLUSTER_CONFIG_PATH" env.
	ClusterConfigPath string `env:"name=KIND_CLUSTER_CONFIG_PATH"`

	// Specify the exported ingress http port for running cluster,
	// configure in "KIND_EXPORT_INGRESS_HTTP_PORT" env,
	// default is created randomly.
	ExportIngressHTTPPort int `env:"name=KIND_EXPORT_INGRESS_HTTP_PORT,fuzzPort"`

	// Specify the exported ingress https port for running cluster,
	// configure in "KIND_EXPORT_INGRESS_HTTPS_PORT" env,
	// default is created randomly.
	ExportIngressHTTPSPort int `env:"name=KIND_EXPORT_INGRESS_HTTPS_PORT,fuzzPort"`
}

func (c *Kind) String() string {
	return fmt.Sprintf("Name: %s, Image: %s, ControlPlanes: %d, Workers: %d", c.ClusterName, c.Image, c.ControlPlanes, c.Workers)
}

func (c *Kind) Startup() error {
	var logger = logs.NewLogger(os.Stdout, 0)
	var provider = cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	// check if the cluster is existed
	var existed, err = isKindClusterExisted(provider, c.ClusterName)
	if err != nil {
		return err
	}

	// remove the existed cluster
	if existed {
		err = provider.Delete(c.ClusterName, "")
		if err != nil {
			return errors.Wrapf(err, "failed to clean the previous cluster")
		}
	}

	// create cluster configuration
	var configOption cluster.CreateOption
	if c.ClusterConfigPath == "" {
		var config, err = renderKindClusterConfiguration(c)
		if err != nil {
			return err
		}
		logger.V(0).Info(string(config))
		configOption = cluster.CreateWithRawConfig(config)
	} else {
		var configPath, err = filepath.Abs(c.ClusterConfigPath)
		if err != nil {
			return errors.Wrapf(err, "failed to load cluster config from path %s", c.ClusterConfigPath)
		}
		configOption = cluster.CreateWithConfigFile(configPath)
	}

	// create cluster
	err = provider.Create(
		c.ClusterName,
		configOption,
		cluster.CreateWithNodeImage(c.Image),
		cluster.CreateWithWaitForReady(c.WaitTimeout),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to startup")
	}
	return nil
}

func (c *Kind) Cleanup() error {
	var logger = logs.NewLogger(os.Stdout, 0)
	var provider = cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	// check if the cluster is existed
	var existed, err = isKindClusterExisted(provider, c.ClusterName)
	if err != nil {
		return err
	}

	if !existed {
		return nil
	}

	// delete cluster
	err = provider.Delete(c.ClusterName, "")
	if err != nil {
		return errors.Wrap(err, "failed to clean")
	}
	return nil
}

func (c *Kind) AddWorker(name string) error {
	return errors.New("add new worker is not supported in kind")
}

func (c *Kind) IsLocalCluster() bool {
	return true
}

func renderKindClusterConfiguration(config *Kind) ([]byte, error) {
	var configTmpl = `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "0.0.0.0"
nodes:
{{- range $index, $element := (until .ControlPlanes) }}
  - role: control-plane
{{- if eq $index 0 }}
    kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "ingress-ready=true"
    extraPortMappings:
    - containerPort: 80
      hostPort: {{ $.ExportIngressHTTPPort }}
      protocol: TCP
    - containerPort: 443
      hostPort: {{ $.ExportIngressHTTPSPort }}
      protocol: TCP
{{- end }}
{{- end }}
{{- range (until .Workers) }}
  - role: worker
{{- end }}
---
`

	var tp, err = template.New("edge").Funcs(sprig.HtmlFuncMap()).Parse(configTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configuration template")
	}
	var output bytes.Buffer
	err = tp.Execute(&output, *config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate configuration")
	}
	return output.Bytes(), nil
}

func isKindClusterExisted(provider *cluster.Provider, clusterName string) (bool, error) {
	var clusters, err = provider.List()
	if err != nil {
		return false, errors.Wrap(err, "failed to list all local clusters")
	}

	for _, cls := range clusters {
		if cls == clusterName {
			return true, nil
		}
	}
	return false, nil
}
