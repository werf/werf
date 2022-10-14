package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/docs/replacers/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
)

var lintCmdData common.CmdData

func NewLintCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub) *cobra.Command {
	cmd := helm.ReplaceHelmLintDocs(helm_v3.NewLintCmd(os.Stdout))

	SetupRenderRelatedWerfChartParams(cmd, &lintCmdData)
	common.SetupEnvironment(&lintCmdData, cmd)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// NOTICE: This is temporary approach to use `werf helm lint` in pipelines correctly — respect --env / WERF_ENV param,
		// NOTICE: which is typically set by the `werf ci-env` command.
		if *lintCmdData.Environment != "" {
			wc.SetStubServiceValuesOverrides(helpers.GetEnvServiceValues(*lintCmdData.Environment))
		}

		if err := InitRenderRelatedWerfChartParams(ctx, &lintCmdData, wc); err != nil {
			return fmt.Errorf("unable to init werf chart: %w", err)
		}

		return oldRunE(cmd, args)
	}

	return cmd
}
