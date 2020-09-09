package helm_v3

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy/werf_chart_v2"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy"
)

func SetupWerfChartForInstallOrUpgradeCmd(cmd *cobra.Command, commonCmdData *cmd_werf_common.CmdData, wc *werf_chart_v2.WerfChart) {
	cmd_werf_common.SetupAddAnnotations(commonCmdData, cmd)
	cmd_werf_common.SetupAddLabels(commonCmdData, cmd)

	cmd_werf_common.SetupSecretValues(commonCmdData, cmd)
	cmd_werf_common.SetupIgnoreSecretKey(commonCmdData, cmd)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		werfChartInitOpts := werf_chart_v2.WerfChartInitOptions{
			SecretValuesFiles: *commonCmdData.SecretValues,
		}
		if extraAnnotations, err := cmd_werf_common.GetUserExtraAnnotations(commonCmdData); err != nil {
			return err
		} else {
			werfChartInitOpts.ExtraAnnotations = extraAnnotations
		}
		if extraLabels, err := cmd_werf_common.GetUserExtraLabels(commonCmdData); err != nil {
			return err
		} else {
			werfChartInitOpts.ExtraLabels = extraLabels
		}

		// NOTE: project-dir is the same as chart-dir for werf helm-v3 install/upgrade commands
		// NOTE: project-dir is werf-project dir only for werf deploy/dismiss commands
		if m, err := deploy.GetSafeSecretManager(context.Background(), "", "", *commonCmdData.SecretValues, *commonCmdData.IgnoreSecretKey); err != nil {
			return err
		} else {
			werfChartInitOpts.SecretsManager = m
		}

		if err := wc.Init("TODO", werfChartInitOpts); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return oldRunE(cmd, args)
	}
}
