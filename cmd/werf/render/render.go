package render

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	RenderOutput string
	Validate     bool
	IncludeCRDs  bool
	ShowOnly     []string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "render [IMAGE_NAME...]",
		Short:                 "Render Kubernetes templates",
		Long:                  common.GetLongCommandDescription(GetRenderDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
			common.DocsLongMD: GetRenderDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			global_warnings.SuppressGlobalWarnings = true
			if *commonCmdData.LogDebug {
				global_warnings.SuppressGlobalWarnings = false
			}
			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runRender(ctx, args) })
		},
	})

	commonCmdData.SetupWithoutImages(cmd)
	commonCmdData.SetupFinalImagesOnly(cmd, true)

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
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
	common.SetupStubTags(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	// TODO(v3): remove, useless
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	// TODO(v3): remove or hide this in all commands, already ignored in v2
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	// TODO(v3): remove, useless for render
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupSkipTLSVerifyKube(&commonCmdData, cmd)
	common.SetupKubeApiServer(&commonCmdData, cmd)
	common.SetupSQLConnectionString(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyHelmDependencies(&commonCmdData, cmd)
	common.SetupKubeCaPath(&commonCmdData, cmd)
	common.SetupKubeTlsServer(&commonCmdData, cmd)
	common.SetupKubeToken(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, true)
	common.SetupNamespace(&commonCmdData, cmd, true)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, true)
	common.SetupSecretValues(&commonCmdData, cmd, true)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupDisableDefaultValues(cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)

	common.SetupKubeVersion(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupKubeQpsLimit(&commonCmdData, cmd)
	common.SetupKubeBurstLimit(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", util.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", util.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"), "Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")

	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"), "Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringArrayVarP(&cmdData.ShowOnly, "show-only", "s", []string{}, "only show manifests rendered from the given templates")

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}

func runRender(ctx context.Context, imageNameListFromArgs []string) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
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

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		commonManager.TerminateSSHAgent()
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

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
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imagesToProcess)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

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
		if err := common.DockerRegistryInit(ctx, &commonCmdData, commonManager.RegistryMirrors()); err != nil {
			return err
		}

		common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
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

		// Override default behavior:
		// Print build logs on error by default.
		// Always print logs if --log-verbose is specified (level.Info).
		isVerbose := logboek.Context(ctx).IsAcceptedLevel(level.Default)
		conveyorOptions.DeferBuildLog = !isVerbose

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if common.GetRequireBuiltImages(ctx, &commonCmdData) {
				shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, imagesToProcess)
				if err != nil {
					return err
				}

				if err := c.ShouldBeBuilt(ctx, shouldBeBuiltOptions); err != nil {
					return err
				}
			} else {
				if err := c.Build(ctx, buildOptions); err != nil {
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
		*commonCmdData.Namespace,
		*commonCmdData.Environment,
		werfConfig,
	)
	if err != nil {
		return fmt.Errorf("get kubernetes namespace: %w", err)
	}

	releaseName, err := deploy_params.GetHelmRelease(
		*commonCmdData.Release,
		*commonCmdData.Environment,
		releaseNamespace,
		werfConfig,
	)
	if err != nil {
		return fmt.Errorf("get helm release name: %w", err)
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
	serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit hash failed: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit time failed: %w", err)
	}
	registryCredentialsPath := docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	serviceValues, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, imagesInfoGetters, helpers.ServiceValuesOptions{
		Namespace:                releaseNamespace,
		Env:                      *commonCmdData.Environment,
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

	file.ChartFileReader = giterminismManager.FileManager

	// TODO(v3): get rid of forcing color mode via ci-env and use color mode detection logic from
	// Nelm instead. Until then, color will be always off here.
	ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultChartRenderLogLevel), action.SetupLoggingOptions{
		ColorMode:      action.LogColorModeOff,
		LogIsParseable: true,
	})
	engine.Debug = *commonCmdData.DebugTemplates

	if _, err := action.ChartRender(ctx, action.ChartRenderOptions{
		ChartAppVersion:              common.GetHelmChartConfigAppVersion(werfConfig),
		ChartDirPath:                 relChartPath,
		ChartRepositoryInsecure:      *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipTLSVerify: *commonCmdData.SkipTlsVerifyHelmDependencies,
		ChartRepositorySkipUpdate:    *commonCmdData.SkipDependenciesRepoRefresh,
		DefaultChartAPIVersion:       chart.APIVersionV2,
		DefaultChartName:             werfConfig.Meta.Project,
		DefaultChartVersion:          "1.0.0",
		DefaultSecretValuesDisable:   *commonCmdData.DisableDefaultSecretValues,
		DefaultValuesDisable:         *commonCmdData.DisableDefaultValues,
		ExtraAnnotations:             extraAnnotations,
		ExtraLabels:                  extraLabels,
		ExtraRuntimeAnnotations:      serviceAnnotations,
		ForceAdoption:                *commonCmdData.ForceAdoption,
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
		LegacyExtraValues:            serviceValues,
		LocalKubeVersion:             *commonCmdData.KubeVersion,
		LogRegistryStreamOut:         os.Stdout,
		NetworkParallelism:           *commonCmdData.NetworkParallelism,
		OutputFilePath:               cmdData.RenderOutput,
		RegistryCredentialsPath:      registryCredentialsPath,
		ReleaseName:                  releaseName,
		ReleaseNamespace:             releaseNamespace,
		ReleaseStorageDriver:         os.Getenv("HELM_DRIVER"),
		SQLConnectionString:  *commonCmdData.SQLConnectionString,
		Remote:                       cmdData.Validate,
		SecretKeyIgnore:              *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:            common.GetSecretValues(&commonCmdData),
		SecretWorkDir:                giterminismManager.ProjectDir(),
		ShowCRDs:                     cmdData.IncludeCRDs,
		ShowOnlyFiles:                append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...),
		ValuesFileSets:               common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:             common.GetValues(&commonCmdData),
		ValuesSets:                   common.GetSet(&commonCmdData),
		ValuesStringSets:             common.GetSetString(&commonCmdData),
	}); err != nil {
		return fmt.Errorf("chart render: %w", err)
	}

	return nil
}
