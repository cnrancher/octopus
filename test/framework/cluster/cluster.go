// +build test

package cluster

import (
	"fmt"
	"io"

	"github.com/rancher/octopus/test/util/exec"
)

type Cluster string

const (
	Kind Cluster = "kind"
	K3d  Cluster = "k3d"
)

func (c Cluster) Startup(rootDir string, writer io.Writer) error {
	if c == "" {
		return nil
	}

	var path = fmt.Sprintf("%s/hack/cluster-%s-startup.sh", rootDir, c)
	return exec.RunBashScript(writer, rootDir, path)
}

func (c Cluster) Cleanup(rootDir string, writer io.Writer) error {
	if c == "" {
		return nil
	}

	var path = fmt.Sprintf("%s/hack/cluster-%s-cleanup.sh", rootDir, c)
	return exec.RunBashScript(writer, rootDir, path)
}

func (c Cluster) AddWorker(rootDir string, writer io.Writer, nodeName string) error {
	if c == "" {
		return nil
	}

	var path = fmt.Sprintf("%s/hack/cluster-%s-addworker.sh", rootDir, c)
	return exec.RunBashScript(writer, rootDir, path, nodeName)
}
