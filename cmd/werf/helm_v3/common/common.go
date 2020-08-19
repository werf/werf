package common

import (
	"fmt"
	"time"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/werf/werf/pkg/deploy/helm_v3"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	helm_v3_cmd "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func SetupCmdActionConfig(cmdData *common.CmdData, cmd *cobra.Command, actionConfig *action.Configuration) {
	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := common.BackgroundContext()

		helm_v3.InitActionConfig(ctx, helm_v3_cmd.Settings, actionConfig, helm_v3.InitActionConfigOptions{
			StatusProgressPeriod:      time.Duration(*cmdData.StatusProgressPeriodSeconds) * time.Second,
			HooksStatusProgressPeriod: time.Duration(*cmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		})

		if err := kube.Init(kube.InitOptions{kube.KubeConfigOptions{
			Context:          *cmdData.KubeContext,
			ConfigPath:       *cmdData.KubeConfig,
			ConfigDataBase64: *cmdData.KubeConfigBase64,
		}}); err != nil {
			return fmt.Errorf("cannot initialize kube: %s", err)
		}

		if err := common.InitKubedog(ctx); err != nil {
			return fmt.Errorf("cannot init kubedog: %s", err)
		}

		if oldRunE != nil {
			return oldRunE(cmd, args)
		}
		return nil
	}

	cmd.Flags().StringVarP(helm_v3_cmd.Settings.GetNamespaceP(), "namespace", "n", *helm_v3_cmd.Settings.GetNamespaceP(), "namespace scope for this request")

	common.SetupKubeConfig(cmdData, cmd)
	common.SetupKubeConfigBase64(cmdData, cmd)
	common.SetupKubeContext(cmdData, cmd)
	common.SetupStatusProgressPeriod(cmdData, cmd)
	common.SetupHooksStatusProgressPeriod(cmdData, cmd)
}
