package helm

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/helm/command_helpers"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/lock_manager"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

var upgradeCmdData cmd_werf_common.CmdData

func NewUpgradeCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub) *cobra.Command {
	postRenderer, err := wc.GetPostRenderer()
	if err != nil {
		panic(err.Error())
	}

	cmd, helmAction := cmd_helm.NewUpgradeCmd(actionConfig, os.Stdout, cmd_helm.UpgradeCmdOptions{
		PostRenderer: postRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &upgradeCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		releaseName := args[0]

		if chartDir, err := helmAction.ChartPathOptions.LocateChart(args[1], cmd_helm.Settings); err != nil {
			return err
		} else {
			if err := InitRenderRelatedWerfChartParams(ctx, &upgradeCmdData, wc, chartDir); err != nil {
				return fmt.Errorf("unable to init werf chart: %s", err)
			}

			if m, err := lock_manager.NewLockManager(cmd_helm.Settings.Namespace()); err != nil {
				return fmt.Errorf("unable to create lock manager: %s", err)
			} else {
				return command_helpers.LockReleaseWrapper(ctx, releaseName, m, func() error {
					return oldRunE(cmd, args)
				})
			}
		}
	}

	return cmd
}
