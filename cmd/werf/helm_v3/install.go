package helm_v3

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/deploy_v2/werf_chart"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

var installCmdData cmd_werf_common.CmdData

func NewInstallCmd(actionConfig *action.Configuration) *cobra.Command {
	wc := werf_chart.NewWerfChart()

	cmd, helmAction := cmd_helm.NewInstallCmd(actionConfig, os.Stdout, cmd_helm.InstallCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
		ValueOpts:    wc.ValueOpts,
	})

	SetupWerfChartParams(cmd, &installCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
			return err
		}

		if err := InitWerfChart(&installCmdData, wc, func(opts *werf_chart.WerfChartInitOptions, args []string) error {
			if releaseName, chartDir, err := helmAction.NameAndChart(args); err != nil {
				return err
			} else {
				opts.ReleaseName = releaseName
				opts.ChartDir = chartDir
			}

			return nil
		}, args); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return wc.WrapInstall(context.Background(), func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
