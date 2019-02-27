package build

import (
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/stages/build/cmd_factory"
)

var CmdData cmd_factory.CmdData
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return cmd_factory.NewCmdWithData(&CmdData, &CommonCmdData)
}
