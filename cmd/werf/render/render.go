package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/config/deploy_params"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	RenderOutput string
	Validate     bool
	IncludeCRDs  bool
	ShowOnly     []string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render",
		Short:                 "Render Kubernetes templates",
		Long:                  common.GetLongCommandDescription(`Render Kubernetes templates. This command will calculate digests and build (if needed) all images defined in the werf.yaml.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.GetContext()

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

			return common.LogRunningTime(func() error { return runRender(ctx) })
		},
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
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: true})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

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

	common.SetupReportPath(&commonCmdData, cmd)
	common.SetupReportFormat(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupSkipBuild(&commonCmdData, cmd)
	common.SetupPlatform(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", common.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", common.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"), "Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")

	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"), "Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringArrayVarP(&cmdData.ShowOnly, "show-only", "s", []string{}, "only show manifests rendered from the given templates")

	return cmd
}

func getShowOnly() []string {
	return append(common.PredefinedValuesByEnvNamePrefix("WERF_SHOW_ONLY"), cmdData.ShowOnly...)
}

func runRender(ctx context.Context) error {
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

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
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

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	namespace, err := deploy_params.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	releaseName, err := deploy_params.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, namespace, werfConfig)
	if err != nil {
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

	buildOptions, err := common.GetBuildOptions(&commonCmdData, giterminismManager, werfConfig)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepository string
	var isStub bool
	var stubImagesNames []string

	if len(werfConfig.StapelImages) != 0 || len(werfConfig.ImagesFromDockerfile) != 0 {
		addr, err := commonCmdData.Repo.GetAddress()
		if err != nil {
			return err
		}

		if addr != storage.LocalStorageAddress {
			if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
				return err
			}

			stagesStorage, err := common.GetStagesStorage(containerBackend, &commonCmdData)
			if err != nil {
				return err
			}
			finalStagesStorage, err := common.GetOptionalFinalStagesStorage(containerBackend, &commonCmdData)
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
			secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerBackend, &commonCmdData)
			if err != nil {
				return err
			}
			cacheStagesStorageList, err := common.GetCacheStagesStorageList(containerBackend, &commonCmdData)
			if err != nil {
				return err
			}

			storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

			imagesRepository = storageManager.StagesStorage.String()

			conveyorOptions, err := common.GetConveyorOptionsWithParallel(&commonCmdData, buildOptions)
			if err != nil {
				return err
			}

			conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, nil, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageLockManager, conveyorOptions)
			defer conveyorWithRetry.Terminate()

			if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
				if *commonCmdData.SkipBuild {
					shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, giterminismManager, werfConfig)
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

				imagesInfoGetters = c.GetImageInfoGetters()

				return nil
			}); err != nil {
				return err
			}

			logboek.LogOptionalLn()
		} else {
			imagesRepository = "REPO"
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

	helmRegistryClientHandler, err := common.NewHelmRegistryClientHandle(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	wc := chart_extender.NewWerfChart(ctx, giterminismManager, secretsManager, chartDir, helm_v3.Settings, helmRegistryClientHandler, chart_extender.WerfChartOptions{
		SecretValueFiles:                  common.GetSecretValues(&commonCmdData),
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
		IgnoreInvalidAnnotationsAndLabels: false,
	})

	if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
		return err
	}
	if err := wc.SetWerfConfig(werfConfig); err != nil {
		return err
	}

	useCustomTagFunc, err := common.GetUseCustomTagFunc(&commonCmdData, giterminismManager, werfConfig)
	if err != nil {
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

	if vals, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, imagesInfoGetters, helpers.ServiceValuesOptions{
		Namespace:                namespace,
		Env:                      *commonCmdData.Environment,
		IsStub:                   isStub,
		StubImagesNames:          stubImagesNames,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
		CustomTagFunc:            useCustomTagFunc,
		CommitHash:               headHash,
		CommitDate:               headTime,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		wc.SetServiceValues(vals)
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, nil, namespace, helm_v3.Settings, helmRegistryClientHandler, actionConfig, helm.InitActionConfigOptions{}); err != nil {
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
		ChartExtender:               wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender { return chart_extender.NewWerfSubchart() },
	}

	templateOpts := helm_v3.TemplateCmdOptions{
		ChainPostRenderer: wc.ChainPostRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   common.GetValues(&commonCmdData),
			StringValues: common.GetSetString(&commonCmdData),
			Values:       common.GetSet(&commonCmdData),
			FileValues:   common.GetSetFile(&commonCmdData),
		},
		Validate:    &cmdData.Validate,
		IncludeCrds: &cmdData.IncludeCRDs,
	}

	fullChartDir := filepath.Join(giterminismManager.ProjectDir(), chartDir)

	if showOnly := getShowOnly(); len(showOnly) > 0 {
		var showFiles []string

		for _, p := range showOnly {
			pAbs := util.GetAbsoluteFilepath(p)
			if strings.HasPrefix(pAbs, fullChartDir) {
				tp := util.GetRelativeToBaseFilepath(fullChartDir, pAbs)
				logboek.Context(ctx).Debug().LogF("Process show-only params: use path %q\n", tp)
				showFiles = append(showFiles, tp)
			} else {
				logboek.Context(ctx).Debug().LogF("Process show-only params: use path %q\n", p)
				showFiles = append(showFiles, p)
			}
		}

		templateOpts.ShowFiles = &showFiles
	}

	helmTemplateCmd, _ := helm_v3.NewTemplateCmd(actionConfig, output, templateOpts)
	if err := helmTemplateCmd.RunE(helmTemplateCmd, []string{releaseName, filepath.Join(giterminismManager.ProjectDir(), chartDir)}); err != nil {
		return fmt.Errorf("helm templates rendering failed: %w", err)
	}

	return nil
}
