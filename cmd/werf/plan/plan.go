package plan

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gookit/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/chrttree"
	helmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/kubeclnt"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/nelm/pkg/resrc"
	"github.com/werf/nelm/pkg/resrcchangcalc"
	"github.com/werf/nelm/pkg/resrcchanglog"
	"github.com/werf/nelm/pkg/resrcpatcher"
	"github.com/werf/nelm/pkg/resrcprocssr"
	"github.com/werf/nelm/pkg/rls"
	"github.com/werf/nelm/pkg/rlsdiff"
	"github.com/werf/nelm/pkg/rlshistor"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/config/deploy_params"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/deploy/helm/maintenance_helper"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
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
	common.SetupDeprecatedReportPath(&commonCmdData, cmd)
	common.SetupDeprecatedReportFormat(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)
	common.SetupSkipBuild(&commonCmdData, cmd)
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
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")
	cmd.Flags().BoolVarP(&cmdData.DetailedExitCode, "exit-code", "", util.GetBoolEnvironmentDefaultFalse("WERF_EXIT_CODE"), "If true, returns exit code 0 if no changes, exit code 2 if any changes planned or exit code 1 in case of an error (default $WERF_EXIT_CODE or false)")

	return cmd
}

func runMain(ctx context.Context, imagesToProcess build.ImagesToProcess) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()
	if !helm.IsExperimentalEngine() {
		global_warnings.GlobalWarningLn(context.Background(), `"werf plan" shows planned changes for the new deployment engine. Currently you are using the old engine, which might produce different results during deployment. If you want to use "werf plan" we strongly advice migrating to the new engine by setting environment variable "WERF_NELM=1".`)
	}

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

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	if *commonCmdData.Follow {
		logboek.LogOptionalLn()
		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager giterminism_manager.Interface) error {
			return run(ctx, containerBackend, headCommitGiterminismManager, imagesToProcess)
		})
	} else {
		return run(ctx, containerBackend, giterminismManager, imagesToProcess)
	}
}

func run(ctx context.Context, containerBackend container_backend.ContainerBackend, giterminismManager giterminism_manager.Interface, imagesToProcess build.ImagesToProcess) error {
	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if err := imagesToProcess.CheckImagesExistence(werfConfig); err != nil {
		return err
	}

	relChartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	chartDir := filepath.Join(giterminismManager.ProjectDir(), relChartDir)

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

	if imagesToProcess.HaveImagesToProcess(werfConfig) {
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

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	namespace, err := deploy_params.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	releaseName, err := deploy_params.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, namespace, werfConfig)
	if err != nil {
		return err
	}

	kubeConfigOptions := kube.KubeConfigOptions{
		Context:          *commonCmdData.KubeContext,
		ConfigPath:       *commonCmdData.KubeConfig,
		ConfigDataBase64: *commonCmdData.KubeConfigBase64,
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %w", err)
	} else {
		lockManager = m
	}

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	wc := chart_extender.NewWerfChart(ctx, giterminismManager, secretsManager, relChartDir, helm_v3.Settings, helmRegistryClient, chart_extender.WerfChartOptions{
		BuildChartDependenciesOpts:        command_helpers.BuildChartDependenciesOptions{SkipUpdate: *commonCmdData.SkipDependenciesRepoRefresh},
		SecretValueFiles:                  common.GetSecretValues(&commonCmdData),
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
		IgnoreInvalidAnnotationsAndLabels: true,
		DisableDefaultValues:              *commonCmdData.DisableDefaultValues,
		DisableDefaultSecretValues:        *commonCmdData.DisableDefaultSecretValues,
	})

	if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
		return err
	}
	if err := wc.SetWerfConfig(werfConfig); err != nil {
		return err
	}

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit hash failed: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit time failed: %w", err)
	}

	if vals, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepo, imagesInfoGetters, helpers.ServiceValuesOptions{
		Namespace:                namespace,
		Env:                      *commonCmdData.Environment,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
		CommitHash:               headHash,
		CommitDate:               headTime,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		wc.SetServiceValues(vals)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender {
			return chart_extender.NewWerfSubchart(ctx, secretsManager, chart_extender.WerfSubchartOptions{
				DisableDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
			})
		},
	}

	// TODO(ilya-lesikov): not needed after new engine migration
	valueOpts := &values.Options{
		ValueFiles:   common.GetValues(&commonCmdData),
		StringValues: common.GetSetString(&commonCmdData),
		Values:       common.GetSet(&commonCmdData),
		FileValues:   common.GetSetFile(&commonCmdData),
	}

	actionConfig, err := common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
	if err != nil {
		return err
	}
	maintenanceHelper := createMaintenanceHelper(ctx, actionConfig, kubeConfigOptions)

	if err := migrateHelm2ToHelm3(ctx, releaseName, namespace, maintenanceHelper, wc.ChainPostRenderer, valueOpts, chartDir, helmRegistryClient); err != nil {
		return err
	}

	actionConfig, err = common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
	if err != nil {
		return err
	}

	if true {
		networkParallelism := common.GetNetworkParallelism(&commonCmdData)
		serviceAnnotations := map[string]string{
			"werf.io/version":      werf.Version,
			"project.werf.io/name": werfConfig.Meta.Project,
			"project.werf.io/env":  *commonCmdData.Environment,
		}

		clientFactory, err := kubeclnt.NewClientFactory()
		if err != nil {
			return fmt.Errorf("error creating kube client factory: %w", err)
		}

		releaseNamespace := resrc.NewReleaseNamespace(&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": lo.WithoutEmpty([]string{namespace, helm_v3.Settings.Namespace()})[0],
				},
			},
		}, resrc.ReleaseNamespaceOptions{
			Mapper: clientFactory.Mapper(),
		})

		chartPathOptions := action.ChartPathOptions{}
		chartPathOptions.SetRegistryClient(actionConfig.RegistryClient)

		return command_helpers.LockReleaseWrapper(ctx, releaseName, lockManager, func() error {
			log.Default.Info(ctx, color.Style{color.Bold, color.Green}.Render("Planning release")+" %q (namespace: %q)", releaseName, releaseNamespace.Name())

			log.Default.Info(ctx, "Constructing release history")
			history, err := rlshistor.NewHistory(releaseName, releaseNamespace.Name(), actionConfig.Releases, rlshistor.HistoryOptions{
				Mapper:          clientFactory.Mapper(),
				DiscoveryClient: clientFactory.Discovery(),
			})
			if err != nil {
				return fmt.Errorf("error constructing release history: %w", err)
			}

			prevRelease, prevReleaseFound, err := history.LastRelease()
			if err != nil {
				return fmt.Errorf("error getting last deployed release: %w", err)
			}

			_, prevDeployedReleaseFound, err := history.LastDeployedRelease()
			if err != nil {
				return fmt.Errorf("error getting last deployed release: %w", err)
			}

			var newRevision int
			var firstDeployed time.Time
			if prevReleaseFound {
				newRevision = prevRelease.Revision() + 1
				firstDeployed = prevRelease.FirstDeployed()
			} else {
				newRevision = 1
			}

			var deployType helmcommon.DeployType
			if prevReleaseFound && prevDeployedReleaseFound {
				deployType = helmcommon.DeployTypeUpgrade
			} else if prevReleaseFound {
				deployType = helmcommon.DeployTypeInstall
			} else {
				deployType = helmcommon.DeployTypeInitial
			}

			log.Default.Info(ctx, "Constructing chart tree")
			chartTree, err := chrttree.NewChartTree(
				ctx,
				chartDir,
				releaseName,
				releaseNamespace.Name(),
				newRevision,
				deployType,
				actionConfig,
				chrttree.ChartTreeOptions{
					StringSetValues: valueOpts.StringValues,
					SetValues:       valueOpts.Values,
					FileValues:      valueOpts.FileValues,
					ValuesFiles:     valueOpts.ValueFiles,
					Mapper:          clientFactory.Mapper(),
					DiscoveryClient: clientFactory.Discovery(),
				},
			)
			if err != nil {
				return fmt.Errorf("error constructing chart tree: %w", err)
			}

			notes := chartTree.Notes()

			var prevRelGeneralResources []*resrc.GeneralResource
			var prevRelFailed bool
			if prevReleaseFound {
				prevRelGeneralResources = prevRelease.GeneralResources()
				prevRelFailed = prevRelease.Failed()
			}

			// FIXME(ilya-lesikov): no releasable resources needed
			log.Default.Info(ctx, "Processing resources")
			resProcessor := resrcprocssr.NewDeployableResourcesProcessor(
				deployType,
				releaseName,
				releaseNamespace,
				chartTree.StandaloneCRDs(),
				chartTree.HookResources(),
				chartTree.GeneralResources(),
				prevRelGeneralResources,
				clientFactory.KubeClient(),
				clientFactory.Mapper(),
				clientFactory.Discovery(),
				resrcprocssr.DeployableResourcesProcessorOptions{
					NetworkParallelism: networkParallelism,
					ReleasableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
						resrcpatcher.NewExtraMetadataPatcher(lo.Assign(userExtraAnnotations, serviceAnnotations), userExtraLabels),
					},
					ReleasableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
						resrcpatcher.NewExtraMetadataPatcher(lo.Assign(userExtraAnnotations, serviceAnnotations), userExtraLabels),
					},
					DeployableStandaloneCRDsPatchers: []resrcpatcher.ResourcePatcher{
						resrcpatcher.NewExtraMetadataPatcher(lo.Assign(userExtraAnnotations, serviceAnnotations), userExtraLabels),
					},
					DeployableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
						resrcpatcher.NewExtraMetadataPatcher(lo.Assign(userExtraAnnotations, serviceAnnotations), userExtraLabels),
					},
					DeployableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
						resrcpatcher.NewExtraMetadataPatcher(lo.Assign(userExtraAnnotations, serviceAnnotations), userExtraLabels),
					},
				},
			)

			if err := resProcessor.Process(ctx); err != nil {
				return fmt.Errorf("error processing deployable resources: %w", err)
			}

			log.Default.Info(ctx, "Constructing new release")
			newRel, err := rls.NewRelease(releaseName, releaseNamespace.Name(), newRevision, chartTree.ReleaseValues(), chartTree.LegacyChart(), resProcessor.ReleasableHookResources(), resProcessor.ReleasableGeneralResources(), notes, rls.ReleaseOptions{
				FirstDeployed: firstDeployed,
				Mapper:        clientFactory.Mapper(),
			})
			if err != nil {
				return fmt.Errorf("error constructing new release: %w", err)
			}

			log.Default.Info(ctx, "Calculating planned changes")
			createdChanges, recreatedChanges, updatedChanges, appliedChanges, deletedChanges, planChangesPlanned := resrcchangcalc.CalculatePlannedChanges(
				releaseName,
				resProcessor.DeployableReleaseNamespaceInfo(),
				resProcessor.DeployableStandaloneCRDsInfos(),
				resProcessor.DeployableHookResourcesInfos(),
				resProcessor.DeployableGeneralResourcesInfos(),
				resProcessor.DeployablePrevReleaseGeneralResourcesInfos(),
				prevRelFailed,
			)

			var releaseUpToDate bool
			if prevReleaseFound {
				releaseUpToDate, err = rlsdiff.ReleaseUpToDate(prevRelease, newRel)
				if err != nil {
					return fmt.Errorf("error checking if release is up to date: %w", err)
				}
			}

			resrcchanglog.LogPlannedChanges(
				ctx,
				releaseName,
				releaseNamespace.Name(),
				!releaseUpToDate,
				createdChanges,
				recreatedChanges,
				updatedChanges,
				appliedChanges,
				deletedChanges,
			)

			if cmdData.DetailedExitCode && (planChangesPlanned || !releaseUpToDate) {
				return resrcchangcalc.ErrChangesPlanned
			}

			return nil
		})
	}

	return nil
}

func createMaintenanceHelper(ctx context.Context, actionConfig *action.Configuration, kubeConfigOptions kube.KubeConfigOptions) *maintenance_helper.MaintenanceHelper {
	maintenanceOpts := maintenance_helper.MaintenanceHelperOptions{
		KubeConfigOptions: kubeConfigOptions,
	}

	for _, val := range []string{
		os.Getenv("WERF_HELM2_RELEASE_STORAGE_NAMESPACE"),
		os.Getenv("WERF_HELM_RELEASE_STORAGE_NAMESPACE"),
		os.Getenv("TILLER_NAMESPACE"),
	} {
		if val != "" {
			maintenanceOpts.Helm2ReleaseStorageNamespace = val
			break
		}
	}

	for _, val := range []string{
		os.Getenv("WERF_HELM2_RELEASE_STORAGE_TYPE"),
		os.Getenv("WERF_HELM_RELEASE_STORAGE_TYPE"),
	} {
		if val != "" {
			maintenanceOpts.Helm2ReleaseStorageType = val
			break
		}
	}

	return maintenance_helper.NewMaintenanceHelper(actionConfig, maintenanceOpts)
}

func migrateHelm2ToHelm3(ctx context.Context, releaseName, namespace string, maintenanceHelper *maintenance_helper.MaintenanceHelper, chainPostRenderer func(postrender.PostRenderer) postrender.PostRenderer, valueOpts *values.Options, fullChartDir string, helmRegistryClient *registry.Client) error {
	if helm2Exists, err := checkHelm2AvailableAndReleaseExists(ctx, releaseName, namespace, maintenanceHelper); err != nil {
		return fmt.Errorf("error checking availability of helm 2 and existence of helm 2 release %q: %w", releaseName, err)
	} else if !helm2Exists {
		return nil
	}

	if helm3Exists, err := checkHelm3ReleaseExists(ctx, releaseName, namespace, maintenanceHelper); err != nil {
		return fmt.Errorf("error checking existence of helm 3 release %q: %w", releaseName, err)
	} else if helm3Exists {
		// helm 2 exists and helm 3 exists
		// migration not needed, but we should warn user that some helm 2 release with the same name exists

		logboek.Context(ctx).Warn().LogF("### Helm 2 and helm 3 release %q exists at the same time ###\n", releaseName)
		logboek.Context(ctx).Warn().LogLn()
		logboek.Context(ctx).Warn().LogF("Found existing helm 2 release %q while there is existing helm 3 release %q in the %q namespace!\n", releaseName, releaseName, namespace)
		logboek.Context(ctx).Warn().LogF("werf will continue deploy process into helm 3 release %q in the %q namespace\n", releaseName, namespace)
		logboek.Context(ctx).Warn().LogF("To disable this warning please remove old helm 2 release %q metadata (fox example using: kubectl -n kube-system delete cm RELEASE_NAME.VERSION)\n", releaseName)
		logboek.Context(ctx).Warn().LogLn()

		return nil
	}

	logboek.Context(ctx).Warn().LogFDetails("Found existing helm 2 release %q, will try to render helm 3 templates and migrate existing release resources to helm 3\n", releaseName)

	logboek.Context(ctx).Default().LogOptionalLn()
	if err := logboek.Context(ctx).LogProcess("Rendering helm 3 templates for the current project state").DoError(func() error {
		actionConfig, err := common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
		if err != nil {
			return err
		}

		helmTemplateCmd, _ := helm_v3.NewTemplateCmd(actionConfig, ioutil.Discard, helm_v3.TemplateCmdOptions{
			StagesSplitter:    helm.NewStagesSplitter(),
			ChainPostRenderer: chainPostRenderer,
			ValueOpts:         valueOpts,
			Validate:          common.NewBool(true),
			IncludeCrds:       common.NewBool(true),
			IsUpgrade:         common.NewBool(true),
		})
		return helmTemplateCmd.RunE(helmTemplateCmd, []string{releaseName, fullChartDir})
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Migrating helm 2 release %q to helm 3 in the %q namespace", releaseName, namespace).DoError(func() error {
		if err := maintenance_helper.Migrate2To3(ctx, releaseName, releaseName, namespace, maintenanceHelper); err != nil {
			return fmt.Errorf("error migrating existing helm 2 release %q to helm 3 release %q in the namespace %q: %w", releaseName, releaseName, namespace, err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func checkHelm2AvailableAndReleaseExists(ctx context.Context, releaseName, namespace string, maintenanceHelper *maintenance_helper.MaintenanceHelper) (bool, error) {
	if available, err := maintenanceHelper.CheckHelm2StorageAvailable(ctx); err != nil {
		return false, err
	} else if available {
		foundHelm2Release, err := maintenanceHelper.IsHelm2ReleaseExist(ctx, releaseName)
		if err != nil {
			return false, fmt.Errorf("error checking existence of helm 2 release %q: %w", releaseName, err)
		}

		return foundHelm2Release, nil
	}

	return false, nil
}

func checkHelm3ReleaseExists(ctx context.Context, releaseName, namespace string, maintenanceHelper *maintenance_helper.MaintenanceHelper) (bool, error) {
	foundHelm3Release, err := maintenanceHelper.IsHelm3ReleaseExist(ctx, releaseName)
	if err != nil {
		return false, fmt.Errorf("error checking existence of helm 3 release %q: %w", releaseName, err)
	}

	return foundHelm3Release, nil
}
