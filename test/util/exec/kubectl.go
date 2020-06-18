// +build test

package exec

import (
	"io"
	"os/exec"
)

// RunKubectl runs the kubectl and redirects Stdout/Stderr to writer.
func RunKubectl(reader io.Reader, writer io.Writer, args ...string) error {
	var cmd = exec.Command("kubectl", args...)
	cmd.Stdin = reader
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}
