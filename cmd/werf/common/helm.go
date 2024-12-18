package common

import (
	"context"
	"time"

	registry_legacy "github.com/werf/3p-helm-for-werf-helm/pkg/registry"
	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/3p-helm/pkg/action"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	bundles_registry "github.com/werf/werf/v2/pkg/deploy/bundles/registry"
	"github.com/werf/werf/v2/pkg/deploy/helm"
	"github.com/werf/werf/v2/pkg/docker"
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

func NewHelmRegistryClientWithoutInitForWerfHelm(ctx context.Context) (*registry_legacy.Client, error) {
	return registry_legacy.NewClient(
		registry_legacy.ClientOptDebug(logboek.Context(ctx).Debug().IsAccepted()),
		registry_legacy.ClientOptWriter(logboek.Context(ctx).OutStream()),
	)
}

func InitHelmRegistryClient(registryClient *registry.Client, dockerConfig string, insecureHelmDependencies bool) {
	registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(dockerConfig))(registryClient)
	if insecureHelmDependencies {
		registry.ClientOptPlainHTTP()
	}
}

func InitHelmRegistryClientForWerfHelm(registryClient *registry_legacy.Client, dockerConfig string, insecureHelmDependencies bool) {
	registry_legacy.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(dockerConfig))(registryClient)
	if insecureHelmDependencies {
		registry_legacy.ClientOptPlainHTTP()
	}
}

func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*bundles_registry.Client, error) {
	out := logboek.Context(ctx).OutStream()

	return bundles_registry.NewClient(
		bundles_registry.ClientOptCredentialsFile(docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)),
		bundles_registry.ClientOptDebug(*commonCmdData.LogDebug),
		bundles_registry.ClientOptInsecure(*commonCmdData.InsecureRegistry),
		bundles_registry.ClientOptSkipTlsVerify(*commonCmdData.SkipTlsVerifyRegistry),
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
