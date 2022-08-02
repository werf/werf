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
)

var templateCmdData common.CmdData

func NewTemplateCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub, namespace *string) *cobra.Command {
	cmd, _ := helm_v3.NewTemplateCmd(actionConfig, os.Stdout, helm_v3.TemplateCmdOptions{
		StagesSplitter:              helm.NewStagesSplitter(),
		StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, namespace),
		ChainPostRenderer:           wc.ChainPostRenderer,
	})
	SetupRenderRelatedWerfChartParams(cmd, &templateCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := InitRenderRelatedWerfChartParams(ctx, &templateCmdData, wc); err != nil {
			return fmt.Errorf("unable to init werf chart: %w", err)
		}

		return oldRunE(cmd, args)
	}

	return cmd
}
