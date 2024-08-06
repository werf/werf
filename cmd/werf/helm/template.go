package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/3p-helm/pkg/action"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/helm"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
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
