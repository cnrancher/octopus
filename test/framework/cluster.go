package framework

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type ClusterKind string

const (
	KindCluster ClusterKind = "kind"
	K3dCluster  ClusterKind = "k3d"
)

type LocalCluster interface {
	Startup(rootDir string, writer io.Writer) error
	Cleanup(rootDir string, writer io.Writer) error
	AddWorker(rootDir string, writer io.Writer, nodeName string) error
}

type localCluster struct {
	kind ClusterKind
}

func (c *localCluster) Startup(rootDir string, writer io.Writer) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-startup.sh", rootDir, c.kind)
	if !isScriptExisted(path) {
		return errors.Errorf("%s cluster startup script isn't existed in %s", c.kind, path)
	}

	var cmd = exec.Command("/usr/bin/env", "bash", path)
	cmd.Dir = rootDir
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}

func (c *localCluster) Cleanup(rootDir string, writer io.Writer) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-cleanup.sh", rootDir, c.kind)
	if !isScriptExisted(path) {
		return errors.Errorf("%s cluster cleanup script isn't existed in %s", c.kind, path)
	}

	var cmd = exec.Command("/usr/bin/env", "bash", path)
	cmd.Dir = rootDir
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}

func (c *localCluster) AddWorker(rootDir string, writer io.Writer, nodeName string) error {
	var path = fmt.Sprintf("%s/hack/cluster-%s-addworker.sh", rootDir, c.kind)
	if !isScriptExisted(path) {
		return errors.Errorf("%s cluster cleanup script isn't existed in %s", c.kind, path)
	}

	var cmd = exec.Command("/usr/bin/env", "bash", path, nodeName)
	cmd.Dir = rootDir
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}

func NewLocalCluster(kind ClusterKind) LocalCluster {
	return &localCluster{
		kind: kind,
	}
}

func isScriptExisted(path string) bool {
	var stat, err = os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}
