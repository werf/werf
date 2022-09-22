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
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util"
)

func NewHelmRegistryClient(ctx context.Context, dockerConfig string, insecureHelmDependencies bool) (*registry.Client, error) {
	client, err := NewHelmRegistryClientWithoutInit(ctx)
	if err != nil {
		return nil, err
	}

	InitHelmRegistryClient(client, dockerConfig, insecureHelmDependencies)

	return client, nil
}

func NewHelmRegistryClientWithoutInit(ctx context.Context) (*registry.Client, error) {
	return registry.NewClient(
		registry.ClientOptDebug(logboek.Context(ctx).Debug().IsAccepted()),
		registry.ClientOptWriter(logboek.Context(ctx).OutStream()),
	)
}

func InitHelmRegistryClient(registryClient *registry.Client, dockerConfig string, insecureHelmDependencies bool) {
	registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(dockerConfig))(registryClient)
	registry.ClientOptInsecure(insecureHelmDependencies)(registryClient)
}

func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*bundles_registry.Client, error) {
	debug := logboek.Context(ctx).Debug().IsAccepted()
	insecure := util.GetBoolEnvironmentDefaultFalse("WERF_BUNDLE_INSECURE_REGISTRY")
	skipTlsVerify := util.GetBoolEnvironmentDefaultFalse("WERF_BUNDLE_SKIP_TLS_VERIFY_REGISTRY")
	out := logboek.Context(ctx).OutStream()

	return bundles_registry.NewClient(
		bundles_registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)),
		bundles_registry.ClientOptDebug(debug),
		bundles_registry.ClientOptInsecure(insecure),
		bundles_registry.ClientOptSkipTlsVerify(skipTlsVerify),
		bundles_registry.ClientOptWriter(out),
	)
}

func NewActionConfig(ctx context.Context, kubeInitializer helm.KubeInitializer, namespace string, commonCmdData *CmdData, registryClient *registry.Client) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := helm.InitActionConfig(ctx, kubeInitializer, namespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:             *commonCmdData.KubeContext,
			ConfigPath:          *commonCmdData.KubeConfig,
			ConfigDataBase64:    *commonCmdData.KubeConfigBase64,
			ConfigPathMergeList: *commonCmdData.KubeConfigPathMergeList,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
		RegistryClient:     registryClient,
	}); err != nil {
		return nil, err
	}

	return actionConfig, nil
}
