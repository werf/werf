package common

import (
	"context"
	"time"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/werf/werf/pkg/deploy/helm"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func NewActionConfig(ctx context.Context, kubeInitializer helm.KubeInitializer, namespace string, commonCmdData *CmdData) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := helm.InitActionConfig(ctx, kubeInitializer, namespace, cmd_helm.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:          *commonCmdData.KubeContext,
			ConfigPath:       *commonCmdData.KubeConfig,
			ConfigDataBase64: *commonCmdData.KubeConfigBase64,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
	}); err != nil {
		return nil, err
	}

	return actionConfig, nil
}
