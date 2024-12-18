package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/3p-helm-for-werf-helm/cmd/helm"
	"github.com/werf/3p-helm-for-werf-helm/pkg/action"
	"github.com/werf/nelm-for-werf-helm/pkg/lock_manager"
	"github.com/werf/werf/v2/cmd/werf/common"
	helm "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm"
	chart_extender "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm"
	command_helpers "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/command_helpers_for_werf_helm"
)

var upgradeCmdData common.CmdData

func NewUpgradeCmd(
	actionConfig *action.Configuration,
	wc *chart_extender.WerfChartStub,
	namespace *string,
) *cobra.Command {
	cmd, _ := helm_v3.NewUpgradeCmd(actionConfig, os.Stdout, helm_v3.UpgradeCmdOptions{
		StagesSplitter:              helm.NewStagesSplitter(),
		StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, namespace),
		ChainPostRenderer:           wc.ChainPostRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &upgradeCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		releaseName := args[0]

		if err := InitRenderRelatedWerfChartParams(ctx, &upgradeCmdData, wc); err != nil {
			return fmt.Errorf("unable to init werf chart: %w", err)
		}

		if m, err := lock_manager.NewLockManager(helm_v3.Settings.Namespace(), true, nil, nil); err != nil {
			return fmt.Errorf("unable to create lock manager: %w", err)
		} else {
			return command_helpers.LockReleaseWrapper(ctx, releaseName, m, func() error {
				return oldRunE(cmd, args)
			})
		}
	}

	return cmd
}
