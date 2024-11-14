package render

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/secrets_manager"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/util"
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

	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
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

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", util.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", util.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"), "Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")

	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"), "Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringArrayVarP(&cmdData.ShowOnly, "show-only", "s", []string{}, "only show manifests rendered from the given templates")

	return cmd
}

func runRender(ctx context.Context, imageNameListFromArgs []string) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitWerf:           true,
		InitGitDataManager: true,
		InitManifestCache:  true,
		InitLRUImagesCache: true,
		InitSSHAgent:       true,
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

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameListFromArgs, true, *commonCmdData.WithoutImages)
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
			ProjectName:      projectName,
			ContainerBackend: containerBackend,
			CmdData:          &commonCmdData,
		})
		if err != nil {
			return fmt.Errorf("unable to init storage manager: %w", err)
		}

		imagesRepository = storageManager.GetServiceValuesRepo()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
		if err != nil {
			return err
		}

		// Override default behaviour:
		// Print build logs on error by default.
		// Always print logs if --log-verbose is specified (level.Info).
		isVerbose := logboek.Context(ctx).IsAcceptedLevel(level.Default)
		conveyorOptions.DeferBuildLog = !isVerbose

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
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

	chartPath := filepath.Join(giterminismManager.ProjectDir(), relChartPath)

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

	serviceAnnotations := map[string]string{
		"werf.io/version":      werf.Version,
		"project.werf.io/name": projectName,
		"project.werf.io/env":  *commonCmdData.Environment,
	}

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

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	secrets_manager.WerfHomeDir = werf.GetHomeDir()

	if err := action.Render(ctx, action.RenderOptions{
		ChartDirPath:               chartPath,
		ChartRepositoryInsecure:    *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipUpdate:  *commonCmdData.SkipDependenciesRepoRefresh,
		DefaultSecretValuesDisable: *commonCmdData.DisableDefaultSecretValues,
		DefaultValuesDisable:       *commonCmdData.DisableDefaultValues,
		ExtraAnnotations:           extraAnnotations,
		ExtraLabels:                extraLabels,
		ExtraRuntimeAnnotations:    serviceAnnotations,
		KubeConfigBase64:           *commonCmdData.KubeConfigBase64,
		KubeConfigPaths:            append([]string{*commonCmdData.KubeConfig}, *commonCmdData.KubeConfigPathMergeList...),
		KubeContext:                *commonCmdData.KubeContext,
		Local:                      !cmdData.Validate,
		LocalKubeVersion:           *commonCmdData.KubeVersion,
		LogDebug:                   *commonCmdData.LogDebug,
		LogRegistryStreamOut:       os.Stdout,
		NetworkParallelism:         *commonCmdData.NetworkParallelism,
		RegistryCredentialsPath:    docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig),
		ReleaseName:                releaseName,
		ReleaseNamespace:           releaseNamespace,
		ReleaseStorageDriver:       action.ReleaseStorageDriver(os.Getenv("HELM_DRIVER")),
		OutputFileSave:             cmdData.RenderOutput != "",
		OutputFilePath:             cmdData.RenderOutput,
		SecretKeyIgnore:            *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:          common.GetSecretValues(&commonCmdData),
		ShowCRDs:                   cmdData.IncludeCRDs,
		ShowOnlyFiles:              append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...),
		ValuesFileSets:             common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:           common.GetValues(&commonCmdData),
		ValuesSets:                 common.GetSet(&commonCmdData),
		ValuesStringSets:           common.GetSetString(&commonCmdData),
		LegacyPreRenderHook: func(
			ctx context.Context,
			releaseNamespace string,
			helmRegistryClient *registry.Client,
			secretsManager *secrets_manager.SecretsManager,
			registryCredentialsPath string,
			chartRepositorySkipUpdate bool,
			secretValuesPaths []string,
			extraAnnotations map[string]string,
			extraLabels map[string]string,
			defaultValuesDisable bool,
			defaultSecretValuesDisable bool,
			helmSettings *cli.EnvSettings,
		) error {
			wc := chart_extender.NewWerfChart(
				ctx,
				giterminismManager,
				secretsManager,
				relChartPath,
				helmSettings,
				helmRegistryClient,
				chart_extender.WerfChartOptions{
					BuildChartDependenciesOpts: command_helpers.BuildChartDependenciesOptions{
						SkipUpdate: chartRepositorySkipUpdate,
					},
					SecretValueFiles:                  secretValuesPaths,
					ExtraAnnotations:                  extraAnnotations,
					ExtraLabels:                       extraLabels,
					IgnoreInvalidAnnotationsAndLabels: true,
					DisableDefaultValues:              defaultValuesDisable,
					DisableDefaultSecretValues:        defaultSecretValuesDisable,
				},
			)

			if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
				return fmt.Errorf("set werf chart environment: %w", err)
			}

			if err := wc.SetWerfConfig(werfConfig); err != nil {
				return fmt.Errorf("set werf chart werf config: %w", err)
			}

			headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
			if err != nil {
				return fmt.Errorf("getting HEAD commit hash failed: %w", err)
			}

			headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
			if err != nil {
				return fmt.Errorf("getting HEAD commit time failed: %w", err)
			}

			if vals, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, imagesInfoGetters, helpers.ServiceValuesOptions{
				Namespace:                releaseNamespace,
				Env:                      *commonCmdData.Environment,
				IsStub:                   isStub,
				DisableEnvStub:           true,
				StubImageNameList:        stubImageNameList,
				SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
				DockerConfigPath:         filepath.Dir(registryCredentialsPath),
				CommitHash:               headHash,
				CommitDate:               headTime,
			}); err != nil {
				return fmt.Errorf("get service values: %w", err)
			} else {
				wc.SetServiceValues(vals)
			}

			loader.GlobalLoadOptions = &loader.LoadOptions{
				ChartExtender: wc,
				SubchartExtenderFactoryFunc: func() chart.ChartExtender {
					return chart_extender.NewWerfSubchart(
						ctx,
						secretsManager,
						chart_extender.WerfSubchartOptions{
							DisableDefaultSecretValues: defaultSecretValuesDisable,
						},
					)
				},
			}

			return nil
		},
	}); err != nil {
		return fmt.Errorf("render manifests: %w", err)
	}

	return nil
}
