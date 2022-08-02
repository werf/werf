package common

import (
	"context"
	"time"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/deploy/helm"
)

func NewHelmRegistryClientHandle(ctx context.Context, commonCmdData *CmdData) (*registry.Client, error) {
	return registry.NewClient(
		registry.ClientOptDebug(logboek.Context(ctx).Debug().IsAccepted()),
		registry.ClientOptInsecure(*commonCmdData.InsecureHelmDependencies),
		registry.ClientOptWriter(logboek.Context(ctx).OutStream()),
	)
}

func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*bundles_registry.Client, error) {
	debug := logboek.Context(ctx).Debug().IsAccepted()
	insecure := *commonCmdData.InsecureHelmDependencies
	out := logboek.Context(ctx).OutStream()

	return bundles_registry.NewClient(
		bundles_registry.ClientOptDebug(debug),
		bundles_registry.ClientOptInsecure(insecure),
		bundles_registry.ClientOptWriter(out),
	)
}

func NewActionConfig(ctx context.Context, kubeInitializer helm.KubeInitializer, namespace string, commonCmdData *CmdData, registryClient *registry.Client) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := helm.InitActionConfig(ctx, kubeInitializer, namespace, helm_v3.Settings, registryClient, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:             *commonCmdData.KubeContext,
			ConfigPath:          *commonCmdData.KubeConfig,
			ConfigDataBase64:    *commonCmdData.KubeConfigBase64,
			ConfigPathMergeList: *commonCmdData.KubeConfigPathMergeList,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
	}); err != nil {
		return nil, err
	}

	return actionConfig, nil
}
