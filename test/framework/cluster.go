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
	K3SCluster  ClusterKind = "k3s"
)

type EmbeddedCluster interface {
	Start(rootDir string, writer io.Writer) error
	Stop(rootDir string, writer io.Writer) error
}

type embeddedCluster struct {
	kind ClusterKind
}

func (c *embeddedCluster) Start(rootDir string, writer io.Writer) error {
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

func (c *embeddedCluster) Stop(rootDir string, writer io.Writer) error {
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

func NewEmbeddedCluster(kind ClusterKind) EmbeddedCluster {
	return &embeddedCluster{
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
