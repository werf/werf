package publish

import (
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/images/publish/cmd_factory"
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return cmd_factory.NewCmdWithData(&CommonCmdData)
}
