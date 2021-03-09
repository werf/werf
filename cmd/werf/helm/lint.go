package helm

import (
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

var lintCmdData cmd_werf_common.CmdData

func NewLintCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub) *cobra.Command {
	cmd := cmd_helm.NewLintCmd(os.Stdout)
	SetupRenderRelatedWerfChartParams(cmd, &lintCmdData)

	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		if err := InitRenderRelatedWerfChartParams(ctx, &lintCmdData, wc); err != nil {
			return fmt.Errorf("unable to init werf chart: %s", err)
		}

		return oldRunE(cmd, args)
	}

	return cmd
}
