// +build test

package exec

import (
	"io"
	"os"
	"os/exec"

	"github.com/gopcua/opcua/errors"
)

// RunBashScript runs the bash script and redirects Stdout/Stderr to writer.
func RunBashScript(writer io.Writer, projectDirPath, scriptPath string, args ...string) error {
	if !isScriptExisted(scriptPath) {
		return errors.Errorf("%s script isn't existed", scriptPath)
	}

	var cmd = exec.Command("/usr/bin/env", append([]string{"bash", scriptPath}, args...)...)
	cmd.Dir = projectDirPath
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}

func isScriptExisted(path string) bool {
	var stat, err = os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}
