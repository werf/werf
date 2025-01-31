package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/3p-helm/pkg/werf/secrets"
	"github.com/werf/3p-helm/pkg/werf/secrets/runtimedata"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, false)
	common.SetupSkipTLSVerifyKube(&commonCmdData, cmd)
	common.SetupKubeApiServer(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyHelmDependencies(&commonCmdData, cmd)
	common.SetupKubeCaPath(&commonCmdData, cmd)
	common.SetupKubeTlsServer(&commonCmdData, cmd)
	common.SetupKubeToken(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, false)
	common.SetupSecretValues(&commonCmdData, cmd, false)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupKubeQpsLimit(&commonCmdData, cmd)
	common.SetupKubeBurstLimit(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupRollbackGraphPath(&commonCmdData, cmd)

	common.SetupRenderSubchartNotes(&commonCmdData, cmd)

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

	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitDockerRegistry: true,
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	repoAddress, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return fmt.Errorf("get repo address: %w", err)
	}

	releaseNamespace := common.GetNamespace(&commonCmdData)
	releaseName, err := common.GetRequiredRelease(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release name: %w", err)
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra annotations: %w", err)
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	deployReportPath, err := common.GetDeployReportPath(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get deploy report path: %w", err)
	}

	var logColorMode action.LogColorMode
	if *commonCmdData.LogColorMode == "auto" {
		logColorMode = action.LogColorModeDefault
	} else {
		logColorMode = action.LogColorMode(*commonCmdData.LogColorMode)
	}

	bundlePath := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundlePath)

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("construct bundles registry client: %w", err)
	}

	if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundlePath, bundlesRegistryClient); err != nil {
		return fmt.Errorf("pull bundle: %w", err)
	}

	bundle, err := chart_extender.NewBundle(ctx, bundlePath, chart_extender.BundleOptions{
		SecretValueFiles:                  common.GetSecretValues(&commonCmdData),
		BuildChartDependenciesOpts:        chart.BuildChartDependenciesOptions{},
		IgnoreInvalidAnnotationsAndLabels: true,
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
	})
	if err != nil {
		return fmt.Errorf("construct bundle: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range bundle.GetExtraAnnotations() {
		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") ||
			key == "werf.io/release-channel" {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	if *commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	extraLabels := bundle.GetExtraLabels()

	if err := action.Deploy(ctx, action.DeployOptions{
		AutoRollback:                 cmdData.AutoRollback,
		ChartDirPath:                 bundle.Dir,
		ChartRepositoryInsecure:      *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipTLSVerify: *commonCmdData.SkipTlsVerifyHelmDependencies,
		ChartRepositorySkipUpdate:    *commonCmdData.SkipDependenciesRepoRefresh,
		DeployGraphPath:              common.GetDeployGraphPath(&commonCmdData),
		DeployGraphSave:              common.GetDeployGraphPath(&commonCmdData) != "",
		DeployReportPath:             deployReportPath,
		DeployReportSave:             common.GetSaveDeployReport(&commonCmdData),
		ExtraAnnotations:             extraAnnotations,
		ExtraLabels:                  extraLabels,
		ExtraRuntimeAnnotations:      serviceAnnotations,
		KubeAPIServerName:            *commonCmdData.KubeApiServer,
		KubeBurstLimit:               *commonCmdData.KubeBurstLimit,
		KubeCAPath:                   *commonCmdData.KubeCaPath,
		KubeConfigBase64:             *commonCmdData.KubeConfigBase64,
		KubeConfigPaths:              append([]string{*commonCmdData.KubeConfig}, *commonCmdData.KubeConfigPathMergeList...),
		KubeContext:                  *commonCmdData.KubeContext,
		KubeQPSLimit:                 *commonCmdData.KubeQpsLimit,
		KubeSkipTLSVerify:            *commonCmdData.SkipTlsVerifyKube,
		KubeTLSServerName:            *commonCmdData.KubeTlsServer,
		KubeToken:                    *commonCmdData.KubeToken,
		LogColorMode:                 logColorMode,
		LogDebug:                     *commonCmdData.LogDebug,
		LogRegistryStreamOut:         os.Stdout,
		NetworkParallelism:           common.GetNetworkParallelism(&commonCmdData),
		ProgressTablePrint:           *commonCmdData.StatusProgressPeriodSeconds != -1,
		ProgressTablePrintInterval:   time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		RegistryCredentialsPath:      docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig),
		ReleaseHistoryLimit:          *commonCmdData.ReleasesHistoryMax,
		ReleaseName:                  releaseName,
		ReleaseNamespace:             releaseNamespace,
		ReleaseStorageDriver:         action.ReleaseStorageDriver(os.Getenv("HELM_DRIVER")),
		RollbackGraphPath:            common.GetRollbackGraphPath(&commonCmdData),
		RollbackGraphSave:            common.GetRollbackGraphPath(&commonCmdData) != "",
		SecretKeyIgnore:              *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:            common.GetSecretValues(&commonCmdData),
		SubNotes:                     *commonCmdData.RenderSubchartNotes,
		TrackCreationTimeout:         time.Duration(cmdData.Timeout) * time.Second,
		TrackDeletionTimeout:         time.Duration(cmdData.Timeout) * time.Second,
		TrackReadinessTimeout:        time.Duration(cmdData.Timeout) * time.Second,
		ValuesFileSets:               common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:             common.GetValues(&commonCmdData),
		ValuesSets:                   common.GetSet(&commonCmdData),
		ValuesStringSets:             common.GetSetString(&commonCmdData),
		LegacyPreDeployHook: func(
			ctx context.Context,
			releaseNamespace string,
			helmRegistryClient *registry.Client,
			registryCredentialsPath string,
			chartRepositorySkipUpdate bool,
			secretValuesPaths []string,
			extraAnnotations map[string]string,
			extraLabels map[string]string,
			defaultValuesDisable bool,
			defaultSecretValuesDisable bool,
			helmSettings *cli.EnvSettings,
		) error {
			if vals, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
				Env:                      *commonCmdData.Environment,
				Namespace:                releaseNamespace,
				SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
				DockerConfigPath:         filepath.Dir(registryCredentialsPath),
			}); err != nil {
				return fmt.Errorf("get service values: %w", err)
			} else {
				bundle.SetServiceValues(vals)
			}

			loader.GlobalLoadOptions = &chart.LoadOptions{
				ChartExtender: bundle,
				SecretsRuntimeDataFactoryFunc: func() runtimedata.RuntimeData {
					return secrets.NewSecretsRuntimeData()
				},
			}
			secrets.CoalesceTablesFunc = chartutil.CoalesceTables

			return nil
		},
	}); err != nil {
		return fmt.Errorf("release deploy: %w", err)
	}

	return nil
}
