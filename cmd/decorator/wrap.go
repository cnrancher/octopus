package decorator

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/util/term"
	cliflag "k8s.io/component-base/cli/flag"
)

func Wrap(c *cobra.Command, nfs cliflag.NamedFlagSets) *cobra.Command {
	var fs = c.Flags()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}

	var usageFmt = "Usage:\n  %s\n"
	var cols, _, _ = term.TerminalSize(c.OutOrStdout())
	c.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), nfs, cols)
		return nil
	})
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), nfs, cols)
	})

	return c
}
