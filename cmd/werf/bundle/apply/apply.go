package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/bundles"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "apply",
		Short:                 "Apply bundle into Kubernetes",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and apply it as a helm chart into Kubernetes cluster.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runApply(ctx) })
		},
	})

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")

	defaultTimeout, err := util.GetIntEnvVar("WERF_TIMEOUT")
	if err != nil || defaultTimeout == nil {
		defaultTimeout = new(int64)
	}
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")

	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "auto-rollback", "R", util.GetBoolEnvironmentDefaultFalse("WERF_AUTO_ROLLBACK"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "atomic", "", util.GetBoolEnvironmentDefaultFalse("WERF_ATOMIC"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_ATOMIC by default)")

	return cmd
}

func runApply(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	repoAddress, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	namespace := common.GetNamespace(&commonCmdData)
	releaseName, err := common.GetRequiredRelease(&commonCmdData)
	if err != nil {
		return err
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:             *commonCmdData.KubeContext,
			ConfigPath:          *commonCmdData.KubeConfig,
			ConfigDataBase64:    *commonCmdData.KubeConfigBase64,
			ConfigPathMergeList: *commonCmdData.KubeConfigPathMergeList,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
		RegistryClient:     helmRegistryClient,
	}); err != nil {
		return err
	}

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundleTmpDir)

	if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundleTmpDir, bundlesRegistryClient); err != nil {
		return fmt.Errorf("unable to pull bundle: %w", err)
	}

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %w", err)
	} else {
		lockManager = m
	}

	if *commonCmdData.Environment != "" {
		userExtraAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	bundle, err := chart_extender.NewBundle(ctx, bundleTmpDir, helm_v3.Settings, helmRegistryClient, secretsManager, chart_extender.BundleOptions{
		SecretValueFiles: common.GetSecretValues(&commonCmdData),
		BuildChartDependenciesOpts: command_helpers.BuildChartDependenciesOptions{
			IgnoreInvalidAnnotationsAndLabels: true,
			SkipUpdate:                        *commonCmdData.SkipDependenciesRepoRefresh,
		},
		IgnoreInvalidAnnotationsAndLabels: true,
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
	})
	if err != nil {
		return err
	}

	if vals, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
		Env:                      *commonCmdData.Environment,
		Namespace:                namespace,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		bundle.SetServiceValues(vals)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: bundle,
	}

	var deployReportPath *string
	if common.GetSaveDeployReport(&commonCmdData) {
		if path, err := common.GetDeployReportPath(&commonCmdData); err != nil {
			return fmt.Errorf("unable to get deploy report path: %w", err)
		} else {
			deployReportPath = &path
		}
	}

	helmUpgradeCmd, _ := helm_v3.NewUpgradeCmd(actionConfig, logboek.Context(ctx).OutStream(), helm_v3.UpgradeCmdOptions{
		StagesSplitter:              helm.NewStagesSplitter(),
		StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, &namespace),
		ChainPostRenderer:           bundle.ChainPostRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   common.GetValues(&commonCmdData),
			StringValues: common.GetSetString(&commonCmdData),
			Values:       common.GetSet(&commonCmdData),
			FileValues:   common.GetSetFile(&commonCmdData),
		},
		CreateNamespace:  common.NewBool(true),
		Install:          common.NewBool(true),
		Wait:             common.NewBool(true),
		Atomic:           common.NewBool(cmdData.AutoRollback),
		Timeout:          common.NewDuration(time.Duration(cmdData.Timeout) * time.Second),
		IgnorePending:    common.NewBool(true),
		CleanupOnFail:    common.NewBool(true),
		DeployReportPath: deployReportPath,
	})

	return command_helpers.LockReleaseWrapper(ctx, releaseName, lockManager, func() error {
		return helmUpgradeCmd.RunE(helmUpgradeCmd, []string{releaseName, bundle.Dir})
	})
}
