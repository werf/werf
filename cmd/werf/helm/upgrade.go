package helm

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/deploy/werf_chart"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

var upgradeCmdData cmd_werf_common.CmdData

func NewUpgradeCmd(actionConfig *action.Configuration) *cobra.Command {
	ctx := common.BackgroundContext()

	wc := werf_chart.NewWerfChart(ctx, nil, "", werf_chart.WerfChartOptions{})

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender {
			return werf_chart.NewWerfChart(ctx, nil, "", werf_chart.WerfChartOptions{})
		},
	}

	cmd, helmAction := cmd_helm.NewUpgradeCmd(actionConfig, os.Stdout, cmd_helm.UpgradeCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
	})

	SetupWerfChartParams(cmd, &upgradeCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := werf.Init(*upgradeCmdData.TmpDir, *upgradeCmdData.HomeDir); err != nil {
			return err
		}

		if err := git_repo.Init(); err != nil {
			return err
		}

		if chartDir, err := helmAction.ChartPathOptions.LocateChart(args[1], cmd_helm.Settings); err != nil {
			return err
		} else {
			wc.ChartDir = chartDir
		}
		wc.ReleaseName = args[0]

		if err := InitWerfChartParams(ctx, &upgradeCmdData, wc, wc.ChartDir); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		if vals, err := werf_chart.GetServiceValues(ctx, "PROJECT", "REPO", "NAMESPACE", nil, werf_chart.ServiceValuesOptions{IsStub: true}); err != nil {
			return fmt.Errorf("error creating service values: %s", err)
		} else if err := wc.SetServiceValues(vals); err != nil {
			return err
		}

		return wc.WrapUpgrade(ctx, func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
