package helm

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/lock_manager"

	"github.com/werf/werf/pkg/deploy/werf_chart"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

var installCmdData cmd_werf_common.CmdData

func NewInstallCmd(actionConfig *action.Configuration, wc *werf_chart.WerfChart) *cobra.Command {
	cmd, helmAction := cmd_helm.NewInstallCmd(actionConfig, os.Stdout, cmd_helm.InstallCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &installCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		if releaseName, chartDir, err := helmAction.NameAndChart(args); err != nil {
			return err
		} else {
			wc.ReleaseName = releaseName
			wc.ChartDir = chartDir
		}

		if err := InitRenderRelatedWerfChartParams(ctx, &installCmdData, wc, wc.ChartDir); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		if m, err := lock_manager.NewLockManager(cmd_helm.Settings.Namespace()); err != nil {
			return fmt.Errorf("unable to create lock manager: %s", err)
		} else {
			wc.LockManager = m
		}

		return wc.WrapInstall(ctx, func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
