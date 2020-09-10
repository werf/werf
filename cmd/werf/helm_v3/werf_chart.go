package helm_v3

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/werf_chart_v2"
	"github.com/werf/werf/pkg/werf"
	cmd_helm "helm.sh/helm/v3/cmd/helm"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
)

func SetupWerfChartParams(cmd *cobra.Command, commonCmdData *cmd_werf_common.CmdData) {
	cmd_werf_common.SetupTmpDir(commonCmdData, cmd)
	cmd_werf_common.SetupHomeDir(commonCmdData, cmd)

	cmd_werf_common.SetupAddAnnotations(commonCmdData, cmd)
	cmd_werf_common.SetupAddLabels(commonCmdData, cmd)

	cmd_werf_common.SetupSecretValues(commonCmdData, cmd)
	cmd_werf_common.SetupIgnoreSecretKey(commonCmdData, cmd)
}

func InitWerfChart(commonCmdData *cmd_werf_common.CmdData, wc *werf_chart_v2.WerfChart, setupWerfChartInitOptionsFunc func(opts *werf_chart_v2.WerfChartInitOptions, args []string) error, args []string) error {
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

	if err := setupWerfChartInitOptionsFunc(&werfChartInitOpts, args); err != nil {
		return err
	}

	// NOTE: project-dir is the same as chart-dir for werf helm-v3 install/upgrade commands
	// NOTE: project-dir is werf-project dir only for werf deploy/dismiss commands
	if m, err := deploy.GetSafeSecretManager(context.Background(), werfChartInitOpts.ChartDir, werfChartInitOpts.ChartDir, *commonCmdData.SecretValues, *commonCmdData.IgnoreSecretKey); err != nil {
		return err
	} else {
		werfChartInitOpts.SecretsManager = m
	}

	if m, err := lock_manager.NewLockManager(cmd_helm.Settings.Namespace()); err != nil {
		return fmt.Errorf("unable to create lock manager: %s", err)
	} else {
		werfChartInitOpts.LockManager = m
	}

	return wc.Init(werfChartInitOpts)
}
