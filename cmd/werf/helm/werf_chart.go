package helm

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/nelm-for-werf-helm/pkg/secrets_manager"
	"github.com/werf/werf/v2/cmd/werf/common"
	chart_extender "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm"
)

func SetupRenderRelatedWerfChartParams(cmd *cobra.Command, commonCmdData *common.CmdData) {
	common.SetupAddAnnotations(commonCmdData, cmd)
	common.SetupAddLabels(commonCmdData, cmd)

	lo.Must0(common.SetupSecretValuesFlags(commonCmdData, cmd))
}

func InitRenderRelatedWerfChartParams(
	ctx context.Context,
	commonCmdData *common.CmdData,
	wc *chart_extender.WerfChartStub,
) error {
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

	wc.SetupSecretValueFiles(commonCmdData.SecretValuesFiles)
	// NOTE: project-dir is the same as chart-dir for werf helm install/upgrade commands
	// NOTE: project-dir is werf-project dir only for werf converge/dismiss commands

	wc.SetupSecretsManager(secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{
		DisableSecretsDecryption: commonCmdData.SecretKeyIgnore,
	}))

	return nil
}
