package plan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/secrets_manager"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/util"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Timeout          int
	DetailedExitCode bool
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
				var imagesToProcess build.ImagesToProcess
				if isSpecificImagesEnabled() {
					imagesToProcess = common.GetImagesToProcess(args, *commonCmdData.WithoutImages)
				}

				return runMain(ctx, imagesToProcess)
			})
		},
	})

	if isSpecificImagesEnabled() {
		commonCmdData.SetupWithoutImages(cmd)
	}

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
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	// TODO(3.0): remove this, useless
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

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
	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupDockerServerStoragePath(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)

	defaultTimeout, err := util.GetIntEnvVar("WERF_TIMEOUT")
	if err != nil || defaultTimeout == nil {
		defaultTimeout = new(int64)
	}
	// TODO(3.0): remove this, useless
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")
	cmd.Flags().BoolVarP(&cmdData.DetailedExitCode, "exit-code", "", util.GetBoolEnvironmentDefaultFalse("WERF_EXIT_CODE"), "If true, returns exit code 0 if no changes, exit code 2 if any changes planned or exit code 1 in case of an error (default $WERF_EXIT_CODE or false)")

	return cmd
}

func runMain(ctx context.Context, imagesToProcess build.ImagesToProcess) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	registryMirrors, err := common.GetContainerRegistryMirror(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("get container registry mirrors: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData, registryMirrors)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData, registryMirrors); err != nil {
		return err
	}

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
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
			headCommitGiterminismManager giterminism_manager.Interface,
		) error {
			return run(ctx, containerBackend, headCommitGiterminismManager, imagesToProcess)
		})
	} else {
		return run(ctx, containerBackend, giterminismManager, imagesToProcess)
	}
}

func run(
	ctx context.Context,
	containerBackend container_backend.ContainerBackend,
	giterminismManager giterminism_manager.Interface,
	imagesToProcess build.ImagesToProcess,
) error {
	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if err := werfConfig.CheckThatImagesExist(imagesToProcess.OnlyImages); err != nil {
		return err
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	imageNameList := common.GetImageNameList(imagesToProcess, werfConfig)
	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imageNameList)
	if err != nil {
		return err
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepo string

	if !imagesToProcess.WithoutImages && (len(werfConfig.StapelImages)+len(werfConfig.ImagesFromDockerfile) > 0) {
		stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		logboek.LogOptionalLn()
		synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
		if err != nil {
			return err
		}
		storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
		if err != nil {
			return err
		}
		secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		cacheStagesStorageList, err := common.GetCacheStagesStorageList(ctx, containerBackend, &commonCmdData)
		if err != nil {
			return err
		}
		useCustomTagFunc, err := common.GetUseCustomTagFunc(&commonCmdData, giterminismManager, imageNameList)
		if err != nil {
			return err
		}

		storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

		imagesRepo = storageManager.GetServiceValuesRepo()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if common.GetRequireBuiltImages(ctx, &commonCmdData) {
				shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, imageNameList)
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
		return fmt.Errorf("get helm release: %w", err)
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

	if err := action.Plan(ctx, action.PlanOptions{
		ChartDirPath:               chartPath,
		ChartRepositoryInsecure:    *commonCmdData.InsecureHelmDependencies,
		ChartRepositorySkipUpdate:  *commonCmdData.SkipDependenciesRepoRefresh,
		DefaultSecretValuesDisable: *commonCmdData.DisableDefaultSecretValues,
		DefaultValuesDisable:       *commonCmdData.DisableDefaultValues,
		ErrorIfChangesPlanned:      cmdData.DetailedExitCode,
		ExtraAnnotations:           extraAnnotations,
		ExtraLabels:                extraLabels,
		ExtraRuntimeAnnotations:    serviceAnnotations,
		KubeConfigBase64:           *commonCmdData.KubeConfigBase64,
		KubeConfigPaths:            append(*commonCmdData.KubeConfigPathMergeList, *commonCmdData.KubeConfig),
		KubeContext:                *commonCmdData.KubeContext,
		LogDebug:                   *commonCmdData.LogDebug,
		LogRegistryStreamOut:       os.Stdout,
		NetworkParallelism:         common.GetNetworkParallelism(&commonCmdData),
		RegistryCredentialsPath:    docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig),
		ReleaseName:                releaseName,
		ReleaseNamespace:           releaseNamespace,
		ReleaseStorageDriver:       action.ReleaseStorageDriver(os.Getenv("HELM_DRIVER")),
		SecretKeyIgnore:            *commonCmdData.IgnoreSecretKey,
		SecretValuesPaths:          common.GetSecretValues(&commonCmdData),
		ValuesFileSets:             common.GetSetFile(&commonCmdData),
		ValuesFilesPaths:           common.GetValues(&commonCmdData),
		ValuesSets:                 common.GetSet(&commonCmdData),
		ValuesStringSets:           common.GetSetString(&commonCmdData),
		LegacyPrePlanHook: func(
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
				return fmt.Errorf("get HEAD commit hash: %w", err)
			}

			headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
			if err != nil {
				return fmt.Errorf("get HEAD commit time: %w", err)
			}

			if vals, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepo, imagesInfoGetters, helpers.ServiceValuesOptions{
				Namespace:                releaseNamespace,
				Env:                      *commonCmdData.Environment,
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
		return fmt.Errorf("plan release: %w", err)
	}

	return nil
}
