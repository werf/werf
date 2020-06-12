package build

import (
	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/stages/build/cmd_factory"
)

var cmdData cmd_factory.CmdData
var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return cmd_factory.NewCmdWithData(&cmdData, &commonCmdData)
}
