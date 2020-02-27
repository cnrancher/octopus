package rootdir

import (
	"os"
	"path/filepath"
)

const (
	envRootDir = "ROOT_DIR"
)

// Get returns the dir of the compile environment,
// pointing the root path by `ROOT_DIR`, otherwise it supposes that
// the tested binaries are placed in the `bin` directory.
func Get() string {
	if rootDir := os.Getenv(envRootDir); rootDir != "" {
		return rootDir
	}
	var currDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	var rootDir, _ = filepath.Abs(filepath.Join(currDir, ".."))
	return rootDir
}
