package helm_v3

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/deploy_v2/werf_chart"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

var upgradeCmdData cmd_werf_common.CmdData

func NewUpgradeCmd(actionConfig *action.Configuration) *cobra.Command {
	wc := werf_chart.NewWerfChart()

	cmd, helmAction := cmd_helm.NewUpgradeCmd(actionConfig, os.Stdout, cmd_helm.UpgradeCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
		ValueOpts:    wc.ValueOpts,
	})

	SetupWerfChartParams(cmd, &upgradeCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
			return err
		}

		if err := InitWerfChart(&upgradeCmdData, wc, func(opts *werf_chart.WerfChartInitOptions, args []string) error {
			if chartDir, err := helmAction.ChartPathOptions.LocateChart(args[1], cmd_helm.Settings); err != nil {
				return err
			} else {
				opts.ChartDir = chartDir
			}

			opts.ReleaseName = args[0]

			return nil
		}, args); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return wc.WrapUpgrade(context.Background(), func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
