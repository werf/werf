package common

import (
	"context"
	"time"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/deploy/helm"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func NewHelmRegistryClientHandle(ctx context.Context) (*helm_v3.RegistryClientHandle, error) {
	if registryClient, err := helm_v3.NewRegistryClient(logboek.Context(ctx).Debug().IsAccepted(), logboek.Context(ctx).OutStream()); err != nil {
		return nil, err
	} else {
		return helm_v3.NewRegistryClientHandle(registryClient), nil
	}
}

func NewActionConfig(ctx context.Context, kubeInitializer helm.KubeInitializer, namespace string, commonCmdData *CmdData, registryClientHandle *helm_v3.RegistryClientHandle) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := helm.InitActionConfig(ctx, kubeInitializer, namespace, cmd_helm.Settings, registryClientHandle, actionConfig, helm.InitActionConfigOptions{
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
