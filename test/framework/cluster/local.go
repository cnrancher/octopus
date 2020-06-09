package cluster

import (
	"fmt"
	"io"

	"github.com/rancher/octopus/test/util/exec"
)

type LocalClusterType string

const (
	LocalClusterTypeKind LocalClusterType = "kind"
	LocalClusterTypeK3d  LocalClusterType = "k3d"
)

type LocalCluster interface {
	Startup(rootDir string, writer io.Writer) error
	Cleanup(rootDir string, writer io.Writer) error
	AddWorker(rootDir string, writer io.Writer, nodeName string) error
}

type localCluster struct {
	kind LocalClusterType
}

func (c *localCluster) Startup(rootDir string, writer io.Writer) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-startup.sh", rootDir, c.kind)
	return exec.RunBashScript(writer, rootDir, path)
}

func (c *localCluster) Cleanup(rootDir string, writer io.Writer) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-cleanup.sh", rootDir, c.kind)
	return exec.RunBashScript(writer, rootDir, path)
}

func (c *localCluster) AddWorker(rootDir string, writer io.Writer, nodeName string) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-addworker.sh", rootDir, c.kind)
	return exec.RunBashScript(writer, rootDir, path, nodeName)
}

func NewLocalCluster(kind LocalClusterType) LocalCluster {
	return &localCluster{
		kind: kind,
	}
}
