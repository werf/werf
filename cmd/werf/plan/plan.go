package plan

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/action"
	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/featgate"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	DetailedExitCode       bool
	DiffContextLines       int
	ShowInsignificantDiffs bool
	ShowSensitiveDiffs     bool
	ShowVerboseCRDDiffs    bool
	ShowVerboseDiffs       bool
}

var commonCmdData common.CmdData

func isSpecificImagesEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_CONVERGE_ENABLE_IMAGES_PARAMS")
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)

	var useMsg string
	if isSpecificImagesEnabled() {
		useMsg = "plan [IMAGE_NAME ...]"
	} else {
		useMsg = "plan"
	}

	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   useMsg,
		Short: "Prepare deploy plan and show how resources in a Kubernetes cluster would change on next deploy",
		Long:  common.GetLongCommandDescription(GetPlanDocs().Long),
		Example: `# Prepare and show deploy plan
werf plan --repo registry.mydomain.com/web --env production`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
			common.DocsLongMD: GetPlanDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				var imageNameListFromArgs []string
				if isSpecificImagesEnabled() {
					imageNameListFromArgs = args
				}

				return runMain(ctx, imageNameListFromArgs)
			})
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: true})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	commonCmdData.SetupWithoutImages(cmd)
	commonCmdData.SetupFinalImagesOnly(cmd, true)
	common.SetupStubTags(&commonCmdData, cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	common.SetupProjectName(&commonCmdData, cmd, false)

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupChartRepoConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupValuesFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupSecretValuesFlags(&commonCmdData, cmd))

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupChartProvenanceKeyring(&commonCmdData, cmd)
	common.SetupChartProvenanceStrategy(&commonCmdData, cmd)
	common.SetupDefaultDeletePropagation(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupExtraRuntimeAnnotations(&commonCmdData, cmd)
	common.SetupExtraRuntimeLabels(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd, true)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupNoFinalTrackingFlag(&commonCmdData, cmd)
	common.SetupNoInstallCRDs(&commonCmdData, cmd)
	common.SetupNoRemoveManualChanges(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd, true)
	common.SetupReleaseInfoAnnotations(&commonCmdData, cmd)
	common.SetupReleaseLabel(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd) // TODO(3.0): remove this, useless
	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupTemplatesAllowDNS(&commonCmdData, cmd)
	common.StubSetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.StubSetupStatusProgressPeriod(&commonCmdData, cmd)
	common.StubSetupTrackTimeout(&commonCmdData, cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	var desc string
	if featgate.FeatGateMoreDetailedExitCodeForPlan.Enabled() || featgate.FeatGatePreviewV2.Enabled() {
		desc = "Return exit code 0 if no changes, 1 if error, 2 if resource changes planned, 3 if no resource changes planned, but release still should be installed (default $WERF_EXIT_CODE or false)"
	} else {
		desc = "Return exit code 0 if no changes, 1 if error, 2 if any changes planned (default $WERF_EXIT_CODE or false)"
	}
	cmd.Flags().BoolVarP(&cmdData.DetailedExitCode, "exit-code", "", util.GetBoolEnvironmentDefaultFalse("WERF_EXIT_CODE"), desc)
	cmd.Flags().BoolVarP(&cmdData.ShowInsignificantDiffs, "show-insignificant-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_INSIGNIFICANT_DIFFS"), "Show insignificant diff lines ($WERF_SHOW_INSIGNIFICANT_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowSensitiveDiffs, "show-sensitive-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_SENSITIVE_DIFFS"), "Show sensitive diff lines ($WERF_SHOW_SENSITIVE_DIFFS by default)")
	cmd.Flags().BoolVarP(&cmdData.ShowVerboseCRDDiffs, "show-verbose-crd-diffs", "", util.GetBoolEnvironmentDefaultFalse("WERF_SHOW_VERBOSE_CRD_DIFFS"), "Show verbose CRD diff lines ($WERF_SHOW_VERBOSE_CRD_DIFFS by default)")
	// TODO(v3): get rid?
	cmd.Flags().BoolVarP(&cmdData.ShowVerboseDiffs, "show-verbose-diffs", "", util.GetBoolEnvironmentDefaultTrue("WERF_SHOW_VERBOSE_DIFFS"), "Show verbose diff lines ($WERF_SHOW_VERBOSE_DIFFS by default)")

	var defaultDiffLines int
	if lines := lo.Must(util.GetIntEnvVar("WERF_DIFF_CONTEXT_LINES")); lines != nil {
		defaultDiffLines = int(*lines)
	} else {
		defaultDiffLines = nelmcommon.DefaultDiffContextLines
	}
	cmd.Flags().IntVarP(&cmdData.DiffContextLines, "diff-context-lines", "", defaultDiffLines, "Show N lines of context around diffs ($WERF_DIFF_CONTEXT_LINES by default)")

	return cmd
}

func runMain(ctx context.Context, imageNameListFromArgs []string) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitManifestCache:           true,
		InitLRUImagesCache:          true,
		InitSSHAgent:                true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	defer func() {
		commonManager.TerminateSSHAgent()
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if *commonCmdData.Follow {
		logboek.LogOptionalLn()
		return common.FollowGitHead(ctx, &commonCmdData, func(
			ctx context.Context,
			headCommitGiterminismManager *giterminism_manager.Manager,
		) error {
			return run(ctx, containerBackend, headCommitGiterminismManager, imageNameListFromArgs)
		})
	} else {
		return run(ctx, containerBackend, giterminismManager, imageNameListFromArgs)
	}
}

func run(
	ctx context.Context,
	containerBackend container_backend.ContainerBackend,
	giterminismManager *giterminism_manager.Manager,
	imageNameListFromArgs []string,
) error {
	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameListFromArgs, *commonCmdData.FinalImagesOnly, *commonCmdData.WithoutImages)
	if err != nil {
		return err
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imagesToProcess)
	if err != nil {
		return err
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepository string
	var isStub bool
	var stubImageNameList []string

	addr, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}

	switch {
	case imagesToProcess.WithoutImages:
	case *commonCmdData.StubTags || addr == storage.LocalStorageAddress:
		imagesRepository = "REPO"
		isStub = true
		stubImageNameList = append(stubImageNameList, imagesToProcess.FinalImageNameList...)
	default:
		logboek.LogOptionalLn()
		common.SetupOndemandKubeInitializer(commonCmdData.KubeContextCurrent, commonCmdData.LegacyKubeConfigPath, commonCmdData.KubeConfigBase64, commonCmdData.LegacyKubeConfigPathsMergeList)
		if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
			return err
		}

		useCustomTagFunc, err := common.GetUseCustomTagFunc(&commonCmdData, giterminismManager, imagesToProcess)
		if err != nil {
			return err
		}

		storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
			ProjectName:                    projectName,
			ContainerBackend:               containerBackend,
			CmdData:                        &commonCmdData,
			CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
			GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		})
		if err != nil {
			return fmt.Errorf("unable to init storage manager: %w", err)
		}

		imagesRepository = storageManager.GetServiceValuesRepo()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if common.GetRequireBuiltImages(&commonCmdData) {
				shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, imagesToProcess)
				if err != nil {
					return err
				}

				if _, err := c.ShouldBeBuilt(ctx, shouldBeBuiltOptions); err != nil {
					return err
				}
			} else {
				if _, err := c.Build(ctx, buildOptions); err != nil {
					return err
				}
			}

			imagesInfoGetters, err = c.GetImageInfoGetters(image.InfoGetterOptions{CustomTagFunc: useCustomTagFunc})
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		logboek.LogOptionalLn()
	}

	relChartPath, err := common.GetHelmChartDir(
		werfConfigPath,
		werfConfig,
		giterminismManager,
	)
	if err != nil {
		return fmt.Errorf("get relative helm chart directory: %w", err)
	}

	releaseNamespace, err := deploy_params.GetKubernetesNamespace(
		commonCmdData.Namespace,
		commonCmdData.Environment,
		werfConfig,
	)
	if err != nil {
		return fmt.Errorf("get kubernetes namespace: %w", err)
	}

	releaseName, err := deploy_params.GetHelmRelease(
		commonCmdData.Release,
		commonCmdData.Environment,
		releaseNamespace,
		werfConfig,
	)
	if err != nil {
		return fmt.Errorf("get helm release: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	if annos, err := common.GetUserExtraAnnotations(&commonCmdData); err != nil {
		return fmt.Errorf("get user extra annotations: %w", err)
	} else {
		for key, value := range annos {
			if strings.HasPrefix(key, "project.werf.io/") ||
				strings.Contains(key, "ci.werf.io/") ||
				key == "werf.io/release-channel" {
				serviceAnnotations[key] = value
			} else {
				extraAnnotations[key] = value
			}
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	serviceAnnotations["project.werf.io/name"] = projectName
	serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment

	extraRuntimeAnnotations := lo.Assign(commonCmdData.ExtraRuntimeAnnotations, serviceAnnotations)
	releaseInfoAnnotations := lo.Assign(commonCmdData.ReleaseInfoAnnotations, serviceAnnotations)

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("get HEAD commit hash: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("get HEAD commit time: %w", err)
	}

	registryCredentialsPath := docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	serviceValues, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, imagesInfoGetters, helpers.ServiceValuesOptions{
		Namespace:                releaseNamespace,
		Env:                      commonCmdData.Environment,
		IsStub:                   isStub,
		DisableEnvStub:           true,
		StubImageNameList:        stubImageNameList,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         filepath.Dir(registryCredentialsPath),
		CommitHash:               headHash,
		CommitDate:               headTime,
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	releaseLabels, err := common.GetReleaseLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release labels: %w", err)
	}

	file.ChartFileReader = giterminismManager.FileManager

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleasePlanInstallLogLevel), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})
	engine.Debug = commonCmdData.DebugTemplates

	if err := action.ReleasePlanInstall(ctx, releaseName, releaseNamespace, action.ReleasePlanInstallOptions{
		KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
		ChartRepoConnectionOptions:  commonCmdData.ChartRepoConnectionOptions,
		ValuesOptions:               commonCmdData.ValuesOptions,
		SecretValuesOptions:         commonCmdData.SecretValuesOptions,
		ChartAppVersion:             common.GetHelmChartConfigAppVersion(werfConfig),
		ChartDirPath:                relChartPath,
		ChartProvenanceKeyring:      commonCmdData.ChartProvenanceKeyring,
		ChartProvenanceStrategy:     commonCmdData.ChartProvenanceStrategy,
		ChartRepoSkipUpdate:         commonCmdData.ChartRepoSkipUpdate,
		DefaultChartAPIVersion:      chart.APIVersionV2,
		DefaultChartName:            werfConfig.Meta.Project,
		DefaultChartVersion:         "1.0.0",
		DefaultDeletePropagation:    commonCmdData.DefaultDeletePropagation,
		DiffContextLines:            cmdData.DiffContextLines,
		ErrorIfChangesPlanned:       cmdData.DetailedExitCode,
		ExtraAnnotations:            extraAnnotations,
		ExtraLabels:                 extraLabels,
		ExtraRuntimeAnnotations:     extraRuntimeAnnotations,
		ExtraRuntimeLabels:          commonCmdData.ExtraRuntimeLabels,
		ForceAdoption:               commonCmdData.ForceAdoption,
		InstallGraphPath:            commonCmdData.InstallGraphPath,
		LegacyExtraValues:           serviceValues,
		LegacyLogRegistryStreamOut:  os.Stdout,
		NetworkParallelism:          commonCmdData.NetworkParallelism,
		NoFinalTracking:             commonCmdData.NoFinalTracking,
		NoInstallStandaloneCRDs:     commonCmdData.NoInstallStandaloneCRDs,
		NoRemoveManualChanges:       commonCmdData.NoRemoveManualChanges,
		RegistryCredentialsPath:     registryCredentialsPath,
		ReleaseInfoAnnotations:      releaseInfoAnnotations,
		ReleaseLabels:               releaseLabels,
		ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
		ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
		ShowInsignificantDiffs:      cmdData.ShowInsignificantDiffs,
		ShowSensitiveDiffs:          cmdData.ShowSensitiveDiffs,
		ShowVerboseCRDDiffs:         cmdData.ShowVerboseCRDDiffs,
		ShowVerboseDiffs:            cmdData.ShowVerboseDiffs,
		TemplatesAllowDNS:           commonCmdData.TemplatesAllowDNS,
	}); err != nil {
		return fmt.Errorf("release plan install: %w", err)
	}

	return nil
}
