package common

import (
	"context"
	"path/filepath"
	"time"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/pkg/homedir"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	bundles_registry "github.com/werf/werf/pkg/deploy/bundles/registry"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/util"
)

func NewHelmRegistryClientHandle(ctx context.Context, commonCmdData *CmdData) (*registry.Client, error) {
	return registry.NewClient(
		registry.ClientOptCredentialsFile(getDockerConfigCredentialsFile(*commonCmdData.DockerConfig)),
		registry.ClientOptDebug(logboek.Context(ctx).Debug().IsAccepted()),
		registry.ClientOptInsecure(*commonCmdData.InsecureHelmDependencies),
		registry.ClientOptWriter(logboek.Context(ctx).OutStream()),
	)
}

func NewBundlesRegistryClient(ctx context.Context, commonCmdData *CmdData) (*bundles_registry.Client, error) {
	debug := logboek.Context(ctx).Debug().IsAccepted()
	insecure := util.GetBoolEnvironmentDefaultFalse("WERF_BUNDLE_INSECURE_REGISTRY")
	skipTlsVerify := util.GetBoolEnvironmentDefaultFalse("WERF_BUNDLE_SKIP_TLS_VERIFY_REGISTRY")
	out := logboek.Context(ctx).OutStream()

	return bundles_registry.NewClient(
		bundles_registry.ClientOptCredentialsFile(getDockerConfigCredentialsFile(*commonCmdData.DockerConfig)),
		bundles_registry.ClientOptDebug(debug),
		bundles_registry.ClientOptInsecure(insecure),
		bundles_registry.ClientOptSkipTlsVerify(skipTlsVerify),
		bundles_registry.ClientOptWriter(out),
	)
}

func getDockerConfigCredentialsFile(configDir string) string {
	if configDir == "" {
		return filepath.Join(homedir.Get(), ".docker", config.ConfigFileName)
	} else {
		return filepath.Join(configDir, config.ConfigFileName)
	}
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
