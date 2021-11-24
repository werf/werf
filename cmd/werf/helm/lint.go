package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
)

var lintCmdData common.CmdData

func NewLintCmd(actionConfig *action.Configuration, wc *chart_extender.WerfChartStub) *cobra.Command {
	cmd := helm_v3.NewLintCmd(os.Stdout)
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
