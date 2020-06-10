package publish

import (
	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	images_publish "github.com/werf/werf/cmd/werf/images/publish/cmd_factory"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return images_publish.NewCmdWithData(&commonCmdData)
}
