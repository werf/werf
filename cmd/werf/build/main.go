package build

import (
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"

	stages_build "github.com/flant/werf/cmd/werf/stages/build"
)

var CmdData stages_build.CmdDataType
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return stages_build.NewCmdWithData(&CmdData, &CommonCmdData)
}
