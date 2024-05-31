package render

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/nelm/pkg/chrttree"
	helmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/kubeclnt"
	"github.com/werf/nelm/pkg/resrc"
	"github.com/werf/nelm/pkg/resrcpatcher"
	"github.com/werf/nelm/pkg/resrcprocssr"
	"github.com/werf/nelm/pkg/rlshistor"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/v2/pkg/deploy/secrets_manager"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/storage/manager"
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
			if *commonCmdData.LogVerbose || *commonCmdData.LogDebug {
				global_warnings.SuppressGlobalWarnings = false
			}
			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runRender(ctx, common.GetImagesToProcess(args, *commonCmdData.WithoutImages)) })
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

	common.SetupSynchronization(&commonCmdData, cmd)

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

func getShowOnly() []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...)
}

func runRender(ctx context.Context, imagesToProcess build.ImagesToProcess) error {
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

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if err := werfConfig.CheckThatImagesExist(imagesToProcess.OnlyImages); err != nil {
		return err
	}

	projectName := werfConfig.Meta.Project

	chartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)

	namespace, err := deploy_params.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	releaseName, err := deploy_params.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, namespace, werfConfig)
	if err != nil {
		return err
	}

	serviceAnnotations := map[string]string{
		"werf.io/version":      werf.Version,
		"project.werf.io/name": werfConfig.Meta.Project,
		"project.werf.io/env":  *commonCmdData.Environment,
	}

	userExtraAnnotations := map[string]string{}
	if annos, err := common.GetUserExtraAnnotations(&commonCmdData); err != nil {
		return err
	} else {
		for key, value := range annos {
			if strings.HasPrefix(key, "project.werf.io/") ||
				strings.Contains(key, "ci.werf.io/") ||
				key == "werf.io/release-channel" {
				serviceAnnotations[key] = value
			} else {
				userExtraAnnotations[key] = value
			}
		}
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	imageNameList := common.GetImageNameList(imagesToProcess, werfConfig)
	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imageNameList)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepo string
	var isStub bool
	var stubImagesNames []string

	if !imagesToProcess.WithoutImages && (len(werfConfig.StapelImages)+len(werfConfig.ImagesFromDockerfile) > 0) {
		addr, err := commonCmdData.Repo.GetAddress()
		if err != nil {
			return err
		}

		if addr != storage.LocalStorageAddress {
			if err := common.DockerRegistryInit(ctx, &commonCmdData, registryMirrors); err != nil {
				return err
			}

			stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
			if err != nil {
				return err
			}
			finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
			if err != nil {
				return err
			}
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

			// Override default behaviour:
			// Print build logs on error by default.
			// Always print logs if --log-verbose is specified (level.Info).
			isVerbose := logboek.Context(ctx).IsAcceptedLevel(level.Default)
			conveyorOptions.DeferBuildLog = !isVerbose

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
		} else {
			imagesRepo = "REPO"
			isStub = true

			for _, img := range werfConfig.StapelImages {
				stubImagesNames = append(stubImagesNames, img.Name)
			}
			for _, img := range werfConfig.ImagesFromDockerfile {
				stubImagesNames = append(stubImagesNames, img.Name)
			}
		}
	}

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	wc := chart_extender.NewWerfChart(ctx, giterminismManager, secretsManager, chartDir, helm_v3.Settings, helmRegistryClient, chart_extender.WerfChartOptions{
		BuildChartDependenciesOpts:        command_helpers.BuildChartDependenciesOptions{SkipUpdate: *commonCmdData.SkipDependenciesRepoRefresh},
		SecretValueFiles:                  common.GetSecretValues(&commonCmdData),
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
		IgnoreInvalidAnnotationsAndLabels: false,
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
		IsStub:                   isStub,
		DisableEnvStub:           true,
		StubImagesNames:          stubImagesNames,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
		CommitHash:               headHash,
		CommitDate:               headTime,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		wc.SetServiceValues(vals)
	}

	actionConfig, err := common.NewActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, &commonCmdData, helmRegistryClient)
	if err != nil {
		return err
	}

	var output io.Writer
	if cmdData.RenderOutput != "" {
		if f, err := os.Create(cmdData.RenderOutput); err != nil {
			return fmt.Errorf("unable to open file %q: %w", cmdData.RenderOutput, err)
		} else {
			defer f.Close()
			output = f
		}
	} else {
		output = os.Stdout
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender {
			return chart_extender.NewWerfSubchart(ctx, secretsManager, chart_extender.WerfSubchartOptions{
				DisableDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
			})
		},
	}

	networkParallelism := common.GetNetworkParallelism(&commonCmdData)

	var clientFactory *kubeclnt.ClientFactory
	if cmdData.Validate {
		clientFactory, err = kubeclnt.NewClientFactory()
		if err != nil {
			return fmt.Errorf("error creating kube client factory: %w", err)
		}
	}

	var releaseNamespaceOptions resrc.ReleaseNamespaceOptions
	if cmdData.Validate {
		releaseNamespaceOptions.Mapper = clientFactory.Mapper()
	}

	releaseNamespace := resrc.NewReleaseNamespace(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": lo.WithoutEmpty([]string{namespace, helm_v3.Settings.Namespace()})[0],
			},
		},
	}, releaseNamespaceOptions)

	// FIXME(ilya-lesikov): there is more chartpath options, are they needed?
	chartPathOptions := action.ChartPathOptions{}
	chartPathOptions.SetRegistryClient(actionConfig.RegistryClient)

	if !cmdData.Validate {
		mem := driver.NewMemory()
		mem.SetNamespace(releaseNamespace.Name())
		actionConfig.Releases = helmstorage.Init(mem)

		actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}
		actionConfig.Capabilities = chartutil.DefaultCapabilities.Copy()

		if *commonCmdData.KubeVersion != "" {
			if kubeVersion, err := chartutil.ParseKubeVersion(*commonCmdData.KubeVersion); err != nil {
				return fmt.Errorf("invalid kube version %q: %w", kubeVersion, err)
			} else {
				actionConfig.Capabilities.KubeVersion = *kubeVersion
			}
		}
	}

	var historyOptions rlshistor.HistoryOptions
	if cmdData.Validate {
		historyOptions.Mapper = clientFactory.Mapper()
		historyOptions.DiscoveryClient = clientFactory.Discovery()
	}

	history, err := rlshistor.NewHistory(releaseName, releaseNamespace.Name(), actionConfig.Releases, historyOptions)
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
	if prevReleaseFound {
		newRevision = prevRelease.Revision() + 1
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

	chartTreeOptions := chrttree.ChartTreeOptions{
		StringSetValues: common.GetSetString(&commonCmdData),
		SetValues:       common.GetSet(&commonCmdData),
		FileValues:      common.GetSetFile(&commonCmdData),
		ValuesFiles:     common.GetValues(&commonCmdData),
	}
	if cmdData.Validate {
		chartTreeOptions.Mapper = clientFactory.Mapper()
		chartTreeOptions.DiscoveryClient = clientFactory.Discovery()
	}

	chartTree, err := chrttree.NewChartTree(
		ctx,
		chartDir,
		releaseName,
		releaseNamespace.Name(),
		newRevision,
		deployType,
		actionConfig,
		chartTreeOptions,
	)
	if err != nil {
		return fmt.Errorf("error constructing chart tree: %w", err)
	}

	var prevRelGeneralResources []*resrc.GeneralResource
	if prevReleaseFound {
		prevRelGeneralResources = prevRelease.GeneralResources()
	}

	resProcessorOptions := resrcprocssr.DeployableResourcesProcessorOptions{
		NetworkParallelism: networkParallelism,
		ReleasableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(userExtraAnnotations, userExtraLabels),
		},
		ReleasableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(userExtraAnnotations, userExtraLabels),
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
	}
	if cmdData.Validate {
		resProcessorOptions.KubeClient = clientFactory.KubeClient()
		resProcessorOptions.Mapper = clientFactory.Mapper()
		resProcessorOptions.DiscoveryClient = clientFactory.Discovery()
		resProcessorOptions.AllowClusterAccess = true
	}

	resProcessor := resrcprocssr.NewDeployableResourcesProcessor(
		deployType,
		releaseName,
		releaseNamespace,
		chartTree.StandaloneCRDs(),
		chartTree.HookResources(),
		chartTree.GeneralResources(),
		prevRelGeneralResources,
		resProcessorOptions,
	)

	if err := resProcessor.Process(ctx); err != nil {
		return fmt.Errorf("error processing deployable resources: %w", err)
	}

	fullChartDir := filepath.Join(giterminismManager.ProjectDir(), chartDir)

	var showFiles []string
	if showOnly := getShowOnly(); len(showOnly) > 0 {
		for _, p := range showOnly {
			pAbs := util.GetAbsoluteFilepath(p)
			if strings.HasPrefix(pAbs, fullChartDir) {
				tp := util.GetRelativeToBaseFilepath(fullChartDir, pAbs)

				if !strings.HasPrefix(tp, chartTree.Name()) {
					tp = filepath.Join(chartTree.Name(), tp)
				}

				logboek.Context(ctx).Debug().LogF("Process show-only params: use path %q\n", tp)
				showFiles = append(showFiles, tp)
			} else {
				if !strings.HasPrefix(p, chartTree.Name()) {
					p = filepath.Join(chartTree.Name(), p)
				}

				logboek.Context(ctx).Debug().LogF("Process show-only params: use path %q\n", p)
				showFiles = append(showFiles, p)
			}
		}
	}

	if cmdData.IncludeCRDs {
		crds := resProcessor.DeployableStandaloneCRDs()

		for _, res := range crds {
			if len(showFiles) > 0 && !lo.Contains(showFiles, res.FilePath()) {
				continue
			}

			if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
				return fmt.Errorf("error rendering CRD %q: %w", res.HumanID(), err)
			}
		}
	}

	hooks := resProcessor.DeployableHookResources()

	for _, res := range hooks {
		if len(showFiles) > 0 && !lo.Contains(showFiles, res.FilePath()) {
			continue
		}

		if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
			return fmt.Errorf("error rendering hook resource %q: %w", res.HumanID(), err)
		}
	}

	resources := resProcessor.DeployableGeneralResources()

	for _, res := range resources {
		if len(showFiles) > 0 && !lo.Contains(showFiles, res.FilePath()) {
			continue
		}

		if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
			return fmt.Errorf("error rendering general resource %q: %w", res.HumanID(), err)
		}
	}

	return nil
}

func renderResource(unstruct *unstructured.Unstructured, path string, output io.Writer) error {
	resourceJsonBytes, err := runtime.Encode(unstructured.UnstructuredJSONScheme, unstruct)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	resourceYamlBytes, err := yaml.JSONToYAML(resourceJsonBytes)
	if err != nil {
		return fmt.Errorf("marshalling to YAML failed: %w", err)
	}

	prefixBytes := []byte(fmt.Sprintf("---\n# Source: %s\n", path))

	if _, err := output.Write(append(prefixBytes, resourceYamlBytes...)); err != nil {
		return fmt.Errorf("writing to output failed: %w", err)
	}

	return nil
}
