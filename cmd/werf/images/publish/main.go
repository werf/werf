package publish

import (
	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/images/publish/cmd_factory"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return cmd_factory.NewCmdWithData(&commonCmdData)
}
