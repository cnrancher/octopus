package verflag

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	ldflagsv "k8s.io/client-go/pkg/version"
)

type versionT struct {
	print  bool
	detail bool
}

var version = versionT{}

// AddFlags registers this package's flags on arbitrary FlagSets, such that they point to the
// same value as the global flags.
func AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&version.print, "version", version.print, "Print version information and quit")
	fs.BoolVar(&version.detail, "full-version", version.detail, "Print full version information and quit")
}

// PrintAndExitIfRequested will check if the -version flag was passed
// and, if so, print the version and exit.
func PrintAndExitIfRequested(name string) {
	if version.detail {
		fmt.Printf("%s %#v\n", name, ldflagsv.Get())
		os.Exit(0)
	}
	if version.print {
		fmt.Printf("%s %s\n", name, ldflagsv.Get().String())
		os.Exit(0)
	}
}
