package docs

import (
	"github.com/spf13/cobra"
)

var CmdData struct {
	dest string
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "docs",
		DisableFlagsInUseLine: true,
		Short:                 "Generate documentation as markdown",
		Hidden:                true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenMarkdownTree(cmd.Root(), CmdData.dest)
		},
	}

	f := cmd.Flags()
	f.StringVar(&CmdData.dest, "dir", "./", "directory to which documentation is written")

	return cmd
}
