package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/deploy/lock_manager"
)

var upgradeCmdData common.CmdData

func NewUpgradeCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub, namespace *string) *cobra.Command {
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

		if m, err := lock_manager.NewLockManager(helm_v3.Settings.Namespace()); err != nil {
			return fmt.Errorf("unable to create lock manager: %w", err)
		} else {
			return command_helpers.LockReleaseWrapper(ctx, releaseName, m, func() error {
				return oldRunE(cmd, args)
			})
		}
	}

	return cmd
}
