package helm_v3

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy_v2/lock_manager"
	"github.com/werf/werf/pkg/deploy_v2/werf_chart"
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

func InitWerfChartParams(commonCmdData *cmd_werf_common.CmdData, wc *werf_chart.WerfChart, chartDir string) error {
	wc.SecretValueFiles = cmd_werf_common.GetSecretValues(commonCmdData)

	if extraAnnotations, err := cmd_werf_common.GetUserExtraAnnotations(commonCmdData); err != nil {
		return err
	} else {
		wc.ExtraAnnotationsAndLabelsPostRenderer.Add(extraAnnotations, nil)
	}

	if extraLabels, err := cmd_werf_common.GetUserExtraLabels(commonCmdData); err != nil {
		return err
	} else {
		wc.ExtraAnnotationsAndLabelsPostRenderer.Add(nil, extraLabels)
	}

	// NOTE: project-dir is the same as chart-dir for werf helm-v3 install/upgrade commands
	// NOTE: project-dir is werf-project dir only for werf deploy/dismiss commands
	if m, err := deploy.GetSafeSecretManager(context.Background(), chartDir, chartDir, cmd_werf_common.GetSecretValues(commonCmdData), *commonCmdData.IgnoreSecretKey); err != nil {
		return err
	} else {
		wc.SecretsManager = m
	}

	if m, err := lock_manager.NewLockManager(cmd_helm.Settings.Namespace()); err != nil {
		return fmt.Errorf("unable to create lock manager: %s", err)
	} else {
		wc.LockManager = m
	}

	return nil
}
