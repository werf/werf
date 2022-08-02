package helm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
)

func SetupRenderRelatedWerfChartParams(cmd *cobra.Command, commonCmdData *common.CmdData) {
	common.SetupAddAnnotations(commonCmdData, cmd)
	common.SetupAddLabels(commonCmdData, cmd)

	common.SetupSecretValues(commonCmdData, cmd)
	common.SetupIgnoreSecretKey(commonCmdData, cmd)
}

func InitRenderRelatedWerfChartParams(ctx context.Context, commonCmdData *common.CmdData, wc *chart_extender.WerfChartStub) error {
	if extraAnnotations, err := common.GetUserExtraAnnotations(commonCmdData); err != nil {
		return err
	} else {
		wc.AddExtraAnnotationsAndLabels(extraAnnotations, nil)
	}

	if extraLabels, err := common.GetUserExtraLabels(commonCmdData); err != nil {
		return err
	} else {
		wc.AddExtraAnnotationsAndLabels(nil, extraLabels)
	}

	wc.SetupSecretValueFiles(common.GetSecretValues(commonCmdData))
	// NOTE: project-dir is the same as chart-dir for werf helm install/upgrade commands
	// NOTE: project-dir is werf-project dir only for werf converge/dismiss commands

	wc.SetupSecretsManager(secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{
		DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey,
	}))

	return nil
}
