package helm_v3

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/deploy"

	"github.com/werf/werf/pkg/deploy/werf_chart_v2"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

func NewInstallCmd(actionConfig *action.Configuration, commonCmdData *cmd_werf_common.CmdData) *cobra.Command {
	wc := werf_chart_v2.NewWerfChart()
	cmd, helmAction := cmd_helm.NewInstallCmd(actionConfig, os.Stdout, cmd_helm.InstallCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
		ValueOpts:    wc.ValueOpts,
	})

	cmd_werf_common.SetupTmpDir(commonCmdData, cmd)
	cmd_werf_common.SetupHomeDir(commonCmdData, cmd)

	cmd_werf_common.SetupAddAnnotations(commonCmdData, cmd)
	cmd_werf_common.SetupAddLabels(commonCmdData, cmd)

	cmd_werf_common.SetupSecretValues(commonCmdData, cmd)
	cmd_werf_common.SetupIgnoreSecretKey(commonCmdData, cmd)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
			return err
		}

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

		_, chartDir, err := helmAction.NameAndChart(args)
		if err != nil {
			return err
		}

		// NOTE: project-dir is the same as chart-dir for werf helm-v3 install/upgrade commands
		// NOTE: project-dir is werf-project dir only for werf deploy/dismiss commands
		if m, err := deploy.GetSafeSecretManager(context.Background(), chartDir, chartDir, *commonCmdData.SecretValues, *commonCmdData.IgnoreSecretKey); err != nil {
			return err
		} else {
			werfChartInitOpts.SecretsManager = m
		}

		if err := wc.Init(chartDir, werfChartInitOpts); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return oldRunE(cmd, args)
	}

	return cmd
}
