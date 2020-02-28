package build

import (
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	stages_build "github.com/flant/werf/cmd/werf/stages/build/cmd_factory"
)

var cmdData stages_build.CmdData
var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return stages_build.NewCmdWithData(&cmdData, &commonCmdData)
}
