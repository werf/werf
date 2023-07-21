package converge

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	helmchart "helm.sh/helm/v3/pkg/werf/chart"
	"helm.sh/helm/v3/pkg/werf/client"
	"helm.sh/helm/v3/pkg/werf/history"
	"helm.sh/helm/v3/pkg/werf/kubeclient"
	"helm.sh/helm/v3/pkg/werf/log"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/plan"
	helmresource "helm.sh/helm/v3/pkg/werf/resource"
	"helm.sh/helm/v3/pkg/werf/resourcebuilder"
	"helm.sh/helm/v3/pkg/werf/resourcetracker"
	"helm.sh/helm/v3/pkg/werf/resourcewaiter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
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
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func isSpecificImagesEnabled() bool {
	return util.GetBoolEnvironmentDefaultFalse("WERF_CONVERGE_ENABLE_IMAGES_PARAMS")
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)

	var useMsg string
	if isSpecificImagesEnabled() {
		useMsg = "converge [IMAGE_NAME ...]"
	} else {
		useMsg = "converge"
	}

	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   useMsg,
		Short: "Build and push images, then deploy application into Kubernetes",
		Long:  common.GetLongCommandDescription(GetConvergeDocs().Long),
		Example: `# Build and deploy current application state into production environment
werf converge --repo registry.mydomain.com/web --env production`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
			common.DocsLongMD: GetConvergeDocs().LongMD,
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
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupDisableDefaultValues(cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)
	common.SetupDeprecatedReportPath(&commonCmdData, cmd)
	common.SetupDeprecatedReportFormat(&commonCmdData, cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

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

	defaultTimeout, err := util.GetIntEnvVar("WERF_TIMEOUT")
	if err != nil || defaultTimeout == nil {
		defaultTimeout = new(int64)
	}
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "auto-rollback", "R", util.GetBoolEnvironmentDefaultFalse("WERF_AUTO_ROLLBACK"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "atomic", "", util.GetBoolEnvironmentDefaultFalse("WERF_ATOMIC"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_ATOMIC by default)")

	return cmd
}

func runMain(ctx context.Context, imagesToProcess build.ImagesToProcess) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData)
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

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

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
	if err := werfConfig.CheckThatImagesExist(imagesToProcess.OnlyImages); err != nil {
		return err
	}

	chartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
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

	wc := chart_extender.NewWerfChart(ctx, giterminismManager, secretsManager, chartDir, helm_v3.Settings, helmRegistryClient, chart_extender.WerfChartOptions{
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

	var extraRuntimeResourceMutators []mutator.RuntimeResourceMutator
	if util.GetBoolEnvironmentDefaultFalse(helm.FEATURE_TOGGLE_ENV_EXPERIMENTAL_DEPLOY_ENGINE) {
		extraRuntimeResourceMutators = []mutator.RuntimeResourceMutator{
			helm.NewExtraAnnotationsMutator(userExtraAnnotations),
			helm.NewExtraLabelsMutator(userExtraLabels),
			helm.NewServiceAnnotationsMutator(*commonCmdData.Environment, werfConfig.Meta.Project),
		}
	}

	actionConfig, err := common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
	if err != nil {
		return err
	}
	maintenanceHelper := createMaintenanceHelper(ctx, actionConfig, kubeConfigOptions)

	if err := migrateHelm2ToHelm3(ctx, releaseName, namespace, maintenanceHelper, wc.ChainPostRenderer, valueOpts, filepath.Join(giterminismManager.ProjectDir(), chartDir), helmRegistryClient, extraRuntimeResourceMutators); err != nil {
		return err
	}

	actionConfig, err = common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
	if err != nil {
		return err
	}

	if util.GetBoolEnvironmentDefaultFalse(helm.FEATURE_TOGGLE_ENV_EXPERIMENTAL_DEPLOY_ENGINE) {
		// FIXME(ilya-lesikov):
		// 1. if last succeeded release was cleaned up because of release limit, werf will see
		// current release as first install. We might want to not delete last succeeded or last
		// uninstalled release ever.
		// 2. rollback should rollback to the last succesfull release, not last release
		// 3. discovery should be without version
		// 4. add adoption validation, whether we can adopt or not

		autoRollback := common.NewBool(cmdData.AutoRollback)
		if *autoRollback {
			return fmt.Errorf("--auto-rollback and --atomic not yet supported with new deploy engine")
		}

		logger := log.NewLogboekLogger(ctx)
		// FIXME(ilya-lesikov): get rid?
		outStream := logboek.OutStream()
		errStream := logboek.ErrStream()

		var deployReportPath string
		if common.GetSaveDeployReport(&commonCmdData) {
			deployReportPath, err = common.GetDeployReportPath(&commonCmdData)
			if err != nil {
				return fmt.Errorf("unable to get deploy report path: %w", err)
			}
		}

		// FIXME(ilya-lesikov): there is more chartpath options, are they needed?
		chartPathOptions := action.ChartPathOptions{}
		chartPathOptions.SetRegistryClient(actionConfig.RegistryClient)

		var releaseNamespace string
		if namespace != "" {
			releaseNamespace = namespace
		} else {
			releaseNamespace = helm_v3.Settings.Namespace()
		}

		configGetter := helm_v3.Settings.GetConfigP()

		deferredKubeClient := kubeclient.NewDeferredKubeClient(*configGetter)
		if err := deferredKubeClient.Init(); err != nil {
			return fmt.Errorf("error initializing deferred kube client: %w", err)
		}

		trackTimeout := *common.NewDuration(time.Duration(cmdData.Timeout) * time.Second)
		waiter := resourcewaiter.NewResourceWaiter(deferredKubeClient.Dynamic(), deferredKubeClient.Mapper(), resourcewaiter.NewResourceWaiterOptions{
			Logger:              logger,
			DefaultTrackTimeout: trackTimeout,
		})

		statusProgressPeriod := time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second
		hooksStatusProgressPeriod := time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second
		tracker := resourcetracker.NewResourceTracker(statusProgressPeriod, hooksStatusProgressPeriod)

		cli, err := client.NewClient(deferredKubeClient.Static(), deferredKubeClient.Dynamic(), deferredKubeClient.Discovery(), deferredKubeClient.Mapper(), waiter)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}
		cli.AddTargetResourceMutators(extraRuntimeResourceMutators...)
		cli.AddTargetResourceMutators(
			mutator.NewReplicasOnCreationMutator(),
			mutator.NewReleaseMetadataMutator(releaseName, namespace),
		)
		cli.SetDeletionTimeout(int(trackTimeout))

		// FIXME(ilya-lesikov): move some of it out of lock release wrapper
		return command_helpers.LockReleaseWrapper(ctx, releaseName, lockManager, func() error {
			actionConfig.Releases.MaxHistory = *commonCmdData.ReleasesHistoryMax

			if err := chartutil.ValidateReleaseName(releaseName); err != nil {
				return fmt.Errorf("release name is invalid: %s", releaseNamespace)
			}

			hist, err := history.NewHistory(releaseName, releaseNamespace, actionConfig.Releases.Driver)
			if err != nil {
				return fmt.Errorf("error building history for release %q: %w", releaseName, err)
			}
			prevRelease := hist.LastRelease()
			deployType := hist.DeployTypeForNextRelease()
			revision := hist.RevisionForNextRelease()

			chartTree, err := helmchart.NewChartTree(chartDir, releaseName, releaseNamespace, revision, deployType, actionConfig, helmchart.NewChartTreeOptions{
				Logger:       logger,
				StringValues: valueOpts.StringValues,
				Values:       valueOpts.Values,
				FileValues:   valueOpts.FileValues,
				ValuesFiles:  valueOpts.ValueFiles,
			})
			if err != nil {
				return fmt.Errorf("error creating chart tree: %w", err)
			}

			releaseNamespace := helmresource.NewUnmanagedResource(
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"metadata": map[string]interface{}{
							"name": releaseNamespace,
						},
					},
				},
				cli.DiscoveryRESTMapper(),
				cli.DiscoveryClient(),
			)
			if err := releaseNamespace.Validate(); err != nil {
				return fmt.Errorf("error validating release namespace: %w", err)
			}

			deployResourceBuilder := resourcebuilder.NewDeployResourceBuilder(releaseNamespace, deployType, cli).
				WithLegacyPreloadedCRDs(chartTree.LegacyPreloadedCRDs()...).
				WithLegacyHelmHooks(chartTree.LegacyHooks()...).
				WithReleaseManifests(chartTree.LegacyResources())

			if prevRelease != nil {
				deployResourceBuilder = deployResourceBuilder.
					WithPrevReleaseManifests(prevRelease.Manifest)
			}

			resources, err := deployResourceBuilder.Build(ctx)
			if err != nil {
				return fmt.Errorf("error building resources: %w", err)
			}

			if os.Getenv("WERF_EXPERIMENTAL_DEPLOY_ENGINE_DEBUG") == "1" {
				for _, res := range resources.HelmResources.UpToDate {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceUpToDateLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceUpToDateLive):\n%s\n", b)
				}
				for _, res := range resources.HelmResources.Outdated {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceOutdatedLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceOutdatedLive):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceOutdatedDesired):\n%s\n", b)
				}
				for _, res := range resources.HelmResources.Unsupported {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceUnsupportedLocal):\n%s\n", b)
				}
				for _, res := range resources.HelmResources.OutdatedImmutable {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceOutdatedImmutableLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceOutdatedImmutableLive):\n%s\n", b)
				}
				for _, res := range resources.HelmResources.NonExisting {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceNonExistingLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmResourceNonExistingDesired):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Matched.UpToDate {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedUpToDateLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedUpToDateLive):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Matched.Outdated {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedOutdatedLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedOutdatedLive):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedOutdatedDesired):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Matched.OutdatedImmutable {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedOutdatedImmutableLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedOutdatedImmutableLive):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Matched.NonExisting {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedNonExistingLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedNonExistingDesired):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Matched.Unsupported {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksMatchedUnsupportedLocal):\n%s\n", b)
				}
				for _, res := range resources.HelmHooks.Unmatched {
					b, _ := json.MarshalIndent(res.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(helmHooksUnmatched):\n%s\n", b)
				}
				for _, res := range resources.PreloadedCRDs.Outdated {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsOutdatedLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsOutdatedLive):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsOutdatedDesired):\n%s\n", b)
				}
				for _, res := range resources.PreloadedCRDs.OutdatedImmutable {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsOutdatedImmutableLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsOutdatedImmutableLive):\n%s\n", b)
				}
				for _, res := range resources.PreloadedCRDs.UpToDate {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsUpToDateLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsUpToDateLive):\n%s\n", b)
				}
				for _, res := range resources.PreloadedCRDs.NonExisting {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsNonExistingLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Desired.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(preloadedCrdsNonExistingDesired):\n%s\n", b)
				}
				for _, res := range resources.PrevReleaseHelmResources.Existing {
					b, _ := json.MarshalIndent(res.Local.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(prevReleaseHelmResourcesExistingLocal):\n%s\n", b)
					b, _ = json.MarshalIndent(res.Live.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(prevReleaseHelmResourcesExistingLive):\n%s\n", b)
				}
				for _, res := range resources.PrevReleaseHelmResources.NonExisting {
					b, _ := json.MarshalIndent(res.Unstructured().UnstructuredContent(), "", "\t")
					fmt.Printf("DEBUG(prevReleaseHelmResourcesNonExistingLocal):\n%s\n", b)
				}
			}

			// FIXME(ilya-lesikov): additional validation here
			// FIXME(ilya-lesikov): move it somewhere?
			var helmHooks []*helmresource.HelmHook
			for _, hook := range resources.HelmHooks.Matched.UpToDate {
				helmHooks = append(helmHooks, hook.Local)
			}
			for _, hook := range resources.HelmHooks.Matched.Outdated {
				helmHooks = append(helmHooks, hook.Local)
			}
			for _, hook := range resources.HelmHooks.Matched.OutdatedImmutable {
				helmHooks = append(helmHooks, hook.Local)
			}
			for _, hook := range resources.HelmHooks.Matched.NonExisting {
				helmHooks = append(helmHooks, hook.Local)
			}
			for _, hook := range resources.HelmHooks.Matched.Unsupported {
				helmHooks = append(helmHooks, hook.Local)
			}

			// FIXME(ilya-lesikov): move it somewhere?
			var helmResources []*helmresource.HelmResource
			for _, res := range resources.HelmResources.UpToDate {
				helmResources = append(helmResources, res.Local)
			}
			for _, res := range resources.HelmResources.Outdated {
				helmResources = append(helmResources, res.Local)
			}
			for _, res := range resources.HelmResources.OutdatedImmutable {
				helmResources = append(helmResources, res.Local)
			}
			for _, res := range resources.HelmResources.NonExisting {
				helmResources = append(helmResources, res.Local)
			}
			for _, res := range resources.HelmResources.Unsupported {
				helmResources = append(helmResources, res.Local)
			}

			rel, err := hist.BuildNextRelease(helmHooks, helmResources, chartTree.Notes(), chartTree.LegacyChart(), chartTree.ReleaseValues())
			if err != nil {
				return fmt.Errorf("error building next release: %w", err)
			}

			succeededRel := hist.PromoteReleaseToSucceeded(rel)

			var supersededRel *release.Release
			if prevRelease != nil {
				supersededRel = hist.PromoteReleaseToSuperseded(prevRelease)
			}

			deployPlan, referencesToCleanupOnFailure, err := plan.
				NewDeployPlanBuilder(deployType, resources.ReleaseNamespace, rel, succeededRel).
				WithPreloadedCRDs(resources.PreloadedCRDs).
				WithMatchedHelmHooks(resources.HelmHooks.Matched).
				WithHelmResources(resources.HelmResources).
				WithPreviousReleaseHelmResources(resources.PrevReleaseHelmResources).
				WithSupersededPreviousRelease(supersededRel).
				WithPreviousReleaseDeployed(hist.LastReleaseIsDeployed()).
				Build(ctx)
			if err != nil {
				return fmt.Errorf("error building deploy plan: %w", err)
			}

			if os.Getenv("WERF_EXPERIMENTAL_DEPLOY_ENGINE_DEBUG") == "1" {
				for _, phase := range deployPlan.Phases {
					fmt.Printf("DEBUG(phaseType): %s\n", phase.Type)
					for _, operation := range phase.Operations {
						fmt.Printf("DEBUG(opType): %s\n", operation.Type())
						switch op := operation.(type) {
						case *plan.OperationCreate:
							for _, target := range op.Targets {
								b, _ := json.MarshalIndent(target.Unstructured().UnstructuredContent(), "", "\t")
								fmt.Printf("DEBUG(opTarget):\n%s\n", b)
							}
						case *plan.OperationUpdate:
							for _, target := range op.Targets {
								b, _ := json.MarshalIndent(target.Unstructured().UnstructuredContent(), "", "\t")
								fmt.Printf("DEBUG(opTarget):\n%s\n", b)
							}
						case *plan.OperationRecreate:
							for _, target := range op.Targets {
								b, _ := json.MarshalIndent(target.Unstructured().UnstructuredContent(), "", "\t")
								fmt.Printf("DEBUG(opTarget):\n%s\n", b)
							}
						case *plan.OperationDelete:
							fmt.Printf("DEBUG(opTargets): %s\n", op.Targets)
						case *plan.OperationCreateReleases:
							for _, r := range op.Releases {
								b, _ := json.MarshalIndent(r, "", "\t")
								fmt.Printf("DEBUG(opTarget):\n%s\n", b)
							}
						case *plan.OperationUpdateReleases:
							for _, r := range op.Releases {
								b, _ := json.MarshalIndent(r, "", "\t")
								fmt.Printf("DEBUG(opTarget):\n%s\n", b)
							}
						}
					}
				}
			}

			if deployPlan.Empty() {
				fmt.Fprintf(outStream, "\nRelease %q in namespace %q canceled: no changes to be made.\n", releaseName, releaseNamespace)
				return nil
			}

			// TODO(ilya-lesikov): add more info from executor report
			if deployReportPath != "" {
				defer func() {
					deployReportData, err := release.NewDeployReport().FromRelease(rel).ToJSONData()
					if err != nil {
						actionConfig.Log("warning: error creating deploy report data: %s", err)
						return
					}

					if err := os.WriteFile(deployReportPath, deployReportData, 0o644); err != nil {
						actionConfig.Log("warning: error writing deploy report file: %s", err)
						return
					}
				}()
			}

			deployReport, executeErr := plan.NewDeployPlanExecutor(deployPlan, releaseNamespace, cli, tracker, actionConfig.Releases).WithTrackTimeout(trackTimeout).Execute(ctx)
			if executeErr != nil {
				defer func() {
					fmt.Fprintf(errStream, "\nRelease %q in namespace %q failed.\n", releaseName, releaseNamespace)
				}()

				rel = hist.PromoteReleaseToFailed(rel)

				finalizeFailedDeployPlan := plan.
					NewFinalizeFailedDeployPlanBuilder(rel).
					WithReferencesToCleanup(referencesToCleanupOnFailure).
					Build()

				// FIXME(ilya-lesikov): deploy report from this execute is not used
				_, err = plan.NewDeployPlanExecutor(finalizeFailedDeployPlan, releaseNamespace, cli, tracker, actionConfig.Releases).WithTrackTimeout(trackTimeout).WithReport(deployReport).Execute(ctx)
				if err != nil {
					return multierror.Append(executeErr, fmt.Errorf("error finalizing failed deploy plan: %w", err))
				}

				return executeErr
			}

			rel = succeededRel

			defer func() {
				fmt.Fprintf(outStream, "\nRelease %q in namespace %q succeeded.\n", releaseName, releaseNamespace)
			}()

			deployReportPrinter := plan.NewDeployReportPrinter(outStream, deployReport)
			deployReportPrinter.PrintSummary()

			// FIXME(ilya-lesikov): better error handling (interrupts, etc)

			// FIXME(ilya-lesikov): don't forget errs.FormatTemplatingError if any errors occurs

			return nil
		})
	} else {
		var deployReportPath *string
		if common.GetSaveDeployReport(&commonCmdData) {
			if path, err := common.GetDeployReportPath(&commonCmdData); err != nil {
				return fmt.Errorf("unable to get deploy report path: %w", err)
			} else {
				deployReportPath = &path
			}
		}

		helmUpgradeCmd, _ := helm_v3.NewUpgradeCmd(actionConfig, logboek.OutStream(), helm_v3.UpgradeCmdOptions{
			StagesSplitter:              helm.NewStagesSplitter(),
			StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, &namespace),
			ChainPostRenderer:           wc.ChainPostRenderer,
			ValueOpts:                   valueOpts,
			CreateNamespace:             common.NewBool(true),
			Install:                     common.NewBool(true),
			Wait:                        common.NewBool(true),
			Atomic:                      common.NewBool(cmdData.AutoRollback),
			Timeout:                     common.NewDuration(time.Duration(cmdData.Timeout) * time.Second),
			IgnorePending:               common.NewBool(true),
			CleanupOnFail:               common.NewBool(true),
			DeployReportPath:            deployReportPath,
		})

		return command_helpers.LockReleaseWrapper(ctx, releaseName, lockManager, func() error {
			if err := helmUpgradeCmd.RunE(helmUpgradeCmd, []string{releaseName, filepath.Join(giterminismManager.ProjectDir(), chartDir)}); err != nil {
				return fmt.Errorf("helm upgrade have failed: %w", err)
			}
			return nil
		})
	}
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

func migrateHelm2ToHelm3(ctx context.Context, releaseName, namespace string, maintenanceHelper *maintenance_helper.MaintenanceHelper, chainPostRenderer func(postrender.PostRenderer) postrender.PostRenderer, valueOpts *values.Options, fullChartDir string, helmRegistryClient *registry.Client, extraMutators []mutator.RuntimeResourceMutator) error {
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
