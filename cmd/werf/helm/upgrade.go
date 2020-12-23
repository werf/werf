package helm

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/werf_chart"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

var upgradeCmdData cmd_werf_common.CmdData

func NewUpgradeCmd(actionConfig *action.Configuration, wc *werf_chart.WerfChart) *cobra.Command {
	cmd, helmAction := cmd_helm.NewUpgradeCmd(actionConfig, os.Stdout, cmd_helm.UpgradeCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &upgradeCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		if chartDir, err := helmAction.ChartPathOptions.LocateChart(args[1], cmd_helm.Settings); err != nil {
			return err
		} else {
			wc.ChartDir = chartDir
		}
		wc.ReleaseName = args[0]

		if err := InitRenderRelatedWerfChartParams(ctx, &upgradeCmdData, wc, wc.ChartDir); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		if m, err := lock_manager.NewLockManager(cmd_helm.Settings.Namespace()); err != nil {
			return fmt.Errorf("unable to create lock manager: %s", err)
		} else {
			wc.LockManager = m
		}

		return wc.WrapUpgrade(ctx, func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
