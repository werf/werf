package docs

import (
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
)

var CmdData struct {
	dest string
}
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "docs",
		DisableFlagsInUseLine: true,
		Short:                 "Generate documentation as markdown",
		Hidden:                true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return GenMarkdownTree(cmd.Root(), CmdData.dest)
		},
	}

	common.SetupLogOptions(&CommonCmdData, cmd)

	f := cmd.Flags()
	f.StringVar(&CmdData.dest, "dir", "./", "directory to which documentation is written")

	return cmd
}
