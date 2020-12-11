package helm

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/werf_chart"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

var templateCmdData cmd_werf_common.CmdData

func NewTemplateCmd(actionConfig *action.Configuration, wc *werf_chart.WerfChart) *cobra.Command {
	cmd, helmAction := cmd_helm.NewTemplateCmd(actionConfig, os.Stdout, cmd_helm.TemplateCmdOptions{
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &templateCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		if releaseName, chartDir, err := helmAction.NameAndChart(args); err != nil {
			return err
		} else {
			wc.ReleaseName = releaseName
			wc.ChartDir = chartDir
		}

		if err := InitRenderRelatedWerfChartParams(ctx, &templateCmdData, wc, wc.ChartDir); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return wc.WrapTemplate(ctx, func() error {
			return oldRunE(cmd, args)
		})
	}

	return cmd
}
