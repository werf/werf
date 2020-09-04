package helm_v3

import (
	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm_v3"
)

func SetupExtraAnnotationsAndLabelsForCmd(cmd *cobra.Command, commonCmdData *cmd_werf_common.CmdData, postRenderer *helm_v3.ExtraAnnotationsAndLabelsPostRenderer) *cobra.Command {
	cmd_werf_common.SetupAddAnnotations(commonCmdData, cmd)
	cmd_werf_common.SetupAddLabels(commonCmdData, cmd)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if extraAnnotations, err := cmd_werf_common.GetUserExtraAnnotations(commonCmdData); err != nil {
			return err
		} else {
			postRenderer.Add(extraAnnotations, nil)
		}

		if extraLabels, err := cmd_werf_common.GetUserExtraLabels(commonCmdData); err != nil {
			return err
		} else {
			postRenderer.Add(nil, extraLabels)
		}

		return oldRunE(cmd, args)
	}

	return cmd
}
