package common

import (
	"context"
	"time"

	"helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/deploy/helm"
)

func NewHelmRegistryClientHandle(ctx context.Context, commonCmdData *CmdData) (*helm_v3.RegistryClientHandle, error) {
	if registryClient, err := helm_v3.NewRegistryClient(logboek.Context(ctx).Debug().IsAccepted(), *commonCmdData.InsecureHelmDependencies, logboek.Context(ctx).OutStream()); err != nil {
		return nil, err
	} else {
		return helm_v3.NewRegistryClientHandle(registryClient), nil
	}
}

func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*registry.Client, error) {
	debug := logboek.Context(ctx).Debug().IsAccepted()
	insecure := *commonCmdData.InsecureHelmDependencies
	out := logboek.Context(ctx).OutStream()

	return registry.NewClient(
		registry.ClientOptDebug(debug),
		registry.ClientOptInsecure(insecure),
		registry.ClientOptWriter(out),
	)
}

func NewActionConfig(ctx context.Context, kubeInitializer helm.KubeInitializer, namespace string, commonCmdData *CmdData, registryClient *helm_v3.RegistryClientHandle) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := helm.InitActionConfig(ctx, kubeInitializer, namespace, helm_v3.Settings, registryClient, actionConfig, helm.InitActionConfigOptions{
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
