package publish

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/nelm/v2/pkg/action"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	helm "github.com/werf/nelm/v2/pkg/helm/pkg/cmd"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/build"
	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/config/deploy_params"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/deploy"
	"github.com/werf/werf/v3/pkg/deploy/bundles"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "publish [IMAGE_NAME...]",
		Short:                 "Publish bundle",
		Long:                  common.GetLongCommandDescription(GetBundlePublishDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
			common.DocsLongMD: GetBundlePublishDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runPublish(ctx, args) })
		},
	})

	commonCmdData.SetupWithoutImages(cmd)
	commonCmdData.SetupFinalImagesOnly(cmd, true)

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

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
	common.SetupMetaRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupLegacyKubeConfigPath(&commonCmdData, cmd)
	common.SetupKubeContextCurrent(&commonCmdData, cmd)
	common.SetupDenoBinaryPath(&commonCmdData, cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)
	common.SetupBuildReportOperations(&commonCmdData, cmd)
	common.SetupUseBuildReport(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	common.SetupProjectName(&commonCmdData, cmd, false)

	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupBackendNetwork(cmd)

	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	lo.Must0(common.SetupChartRepoConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupValuesFlags(&commonCmdData, cmd))
	common.SetupSecretValuesFileFlags(&commonCmdData, cmd)

	common.SetupPatchesFlags(&commonCmdData, cmd, common.SetupPatchesFlagsOptions{})

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Publish bundle into container registry repo by the provided tag and by the chart version: the tag itself if it is strict semver (a leading `v` is dropped), otherwise `0.0.0-<unix time>-<tag slug>`, where the slug is the lowercased tag with each run of characters other than letters and digits replaced by `-`, cut and suffixed with `-<8 hex digits of the tag SHA-256>` if the version would be longer than 128 characters. A `+` in the pushed tags is replaced by `_` ($WERF_TAG or latest by default)")

	return cmd
}

func runPublish(ctx context.Context, imageNameListFromArgs []string) error {
	ctx, logOperationsSummaryFn := common.InitOperationsStatistics(ctx, &commonCmdData)
	defer logOperationsSummaryFn()
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry: true,
		InitWerf:           true,
		InitGitDataManager: true,
		InitManifestCache:  true,
		InitLRUImagesCache: true,
		InitSSHAgent:       true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
		if containerBackend, ok := commonManager.TryContainerBackend(); ok {
			if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
				logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
			}
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

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameListFromArgs, *commonCmdData.FinalImagesOnly, *commonCmdData.WithoutImages)
	if err != nil {
		return err
	}

	var containerBackend container_backend.ContainerBackend
	if !imagesToProcess.WithoutImages {
		var newCtx context.Context
		containerBackend, newCtx, err = commonManager.EnsureContainerBackend(ctx, &commonCmdData, common.EnsureContainerBackendOptions{InitDockerRegistry: true, RequireDockerDaemon: true})
		if err != nil {
			return fmt.Errorf("container backend initialization error: %w", err)
		}
		ctx = newCtx
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

	logboek.LogOptionalLn()

	bundleRepo, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}
	if err := common.ValidateRepoContainerRegistry(commonCmdData.Repo.GetContainerRegistry(ctx)); err != nil {
		return err
	}
	if *commonCmdData.FinalRepo.Address != "" {
		if err := common.ValidateRepoContainerRegistry(commonCmdData.FinalRepo.GetContainerRegistry(ctx)); err != nil {
			return err
		}
		bundleRepo, err = commonCmdData.FinalRepo.GetAddress()
		if err != nil {
			return err
		}
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepo string

	if !imagesToProcess.WithoutImages {

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

		imagesRepo = storageManager.GetServiceValuesRepo()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if c.UseBuildReport {
				logboek.Context(ctx).Debug().LogFDetails("Avoid building because of using build report: %s\n", c.BuildReportPath)

				imagesInfoGetters, err = c.GetImageInfoGettersFromReport(ctx, image.InfoGetterOptions{CustomTagFunc: useCustomTagFunc})
				if err != nil {
					return err
				}
			} else {
				if common.GetRequireBuiltImages(&commonCmdData) {
					shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, werfConfig, imagesToProcess)
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
			}

			return nil
		}); err != nil {
			return err
		}

		logboek.LogOptionalLn()
	}

	chartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	nelmcommon.ChartFileReader = giterminismManager.FileManager

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit hash failed: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit time failed: %w", err)
	}

	serviceValues, err := deploy.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepo, imagesInfoGetters, deploy.ServiceValuesOptions{
		Env:        commonCmdData.Environment,
		CommitHash: headHash,
		CommitDate: headTime,
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	chartPackOpts, err := newChartPackOptions(ctx, werfConfig, serviceValues)
	if err != nil {
		return err
	}

	chartPackOpts.Destination = filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(chartPackOpts.Destination)

	chartPackOpts.RegistryCredentialsPath = docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig)

	ctx = common.SetupDenoContext(ctx)

	return bundles.Publish(ctx, filepath.Join(giterminismManager.ProjectDir(), chartDir), bundleRepo, cmdData.Tag, bundlesRegistryClient, bundles.PublishOptions{
		ChartPackOptions: chartPackOpts,
	})
}

// newChartPackOptions returns the options to pack the chart with the publish-time values, annotations,
// labels, images and the release name and namespace that converge would deploy it with.
func newChartPackOptions(ctx context.Context, werfConfig *config.WerfConfig, serviceValues map[string]interface{}) (action.ChartPackOptions, error) {
	annotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return action.ChartPackOptions{}, fmt.Errorf("get user extra annotations: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range annotations {
		if key == "werf.io/release-channel" {
			continue
		}

		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["project.werf.io/name"] = werfConfig.Meta.Project
	serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return action.ChartPackOptions{}, fmt.Errorf("get user extra labels: %w", err)
	}

	releaseNamespace, err := deploy_params.GetKubernetesNamespace("", commonCmdData.Environment, werfConfig)
	if err != nil {
		return action.ChartPackOptions{}, fmt.Errorf("get kubernetes namespace: %w", err)
	}

	releaseName, err := deploy_params.GetHelmRelease("", commonCmdData.Environment, releaseNamespace, werfConfig)
	if err != nil {
		return action.ChartPackOptions{}, fmt.Errorf("get helm release name: %w", err)
	}

	return action.ChartPackOptions{
		ChartRepoConnectionOptions: commonCmdData.ChartRepoConnectionOptions,
		AppVersion:                 common.GetHelmChartConfigAppVersion(werfConfig),
		ChartRepoSkipUpdate:        commonCmdData.ChartRepoSkipUpdate,
		ContentCache:               helm.Settings.ContentCache,
		DefaultPatchesDisable:      lo.Ternary(commonCmdData.DefaultPatchesDisable, lo.ToPtr(true), nil),
		DefaultSecretValuesDisable: lo.Ternary(commonCmdData.DefaultSecretValuesDisable, lo.ToPtr(true), nil),
		DefaultValuesDisable:       lo.Ternary(commonCmdData.DefaultValuesDisable, lo.ToPtr(true), nil),
		DenoBinaryPath:             commonCmdData.DenoBinaryPath,
		ExtraAnnotations:           extraAnnotations,
		ExtraLabels:                extraLabels,
		ExtraRuntimeAnnotations:    serviceAnnotations,
		ExtraValues:                serviceValues,
		Images:                     bundles.ChartImagesFromServiceValues(ctx, serviceValues),
		PatchesFiles:               commonCmdData.PatchesFiles,
		Release: &werfdata.ReleaseMetadata{
			Name:      releaseName,
			Namespace: releaseNamespace,
			Env:       commonCmdData.Environment,
		},
		ReleaseInfoAnnotations: serviceAnnotations,
		RepositoryCache:        helm.Settings.RepositoryCache,
		RepositoryConfig:       helm.Settings.RepositoryConfig,
		RootSetJSON:            commonCmdData.RootSetJSON,
		SecretValuesFiles:      commonCmdData.SecretValuesFiles,
		TempDirPath:            lo.FromPtr(commonCmdData.TmpDir),
		ValuesFiles:            commonCmdData.ValuesFiles,
		ValuesSet:              commonCmdData.ValuesSet,
		ValuesSetFile:          commonCmdData.ValuesSetFile,
		ValuesSetJSON:          commonCmdData.ValuesSetJSON,
		ValuesSetLiteral:       commonCmdData.ValuesSetLiteral,
		ValuesSetString:        commonCmdData.ValuesSetString,
	}, nil
}
