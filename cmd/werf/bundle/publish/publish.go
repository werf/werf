package publish

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/getter"

	uuid "github.com/satori/go.uuid"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf/global_warnings"

	"github.com/werf/werf/pkg/deploy/helm"

	"github.com/werf/werf/pkg/deploy/secrets_manager"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Tag string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish bundle",
		Long: common.GetLongCommandDescription(`Publish bundle into the container registry. Werf bundle contains built images defined in the werf.yaml, helm chart, service values which contain built images tags, any custom values and set values params provided during publish invocation, werf addon templates (like werf_image).

Published into container registry bundle can be rolled out by the "werf bundle" command.
`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

			logboek.Streams().Mute()
			logboek.SetAcceptedLevel(level.Error)

			if err := common.ProcessLogOptionsDefaultQuiet(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(runPublish)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupGiterminismInspectorOptions(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	common.SetupReportPath(&commonCmdData, cmd)
	common.SetupReportFormat(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(&commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupSkipBuild(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Publish bundle into container registry repo by the provided tag ($WERF_TAG or latest by default)")

	return cmd
}

func runPublish() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(ctx, *commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	giterminismManager, err := common.GetGiterminismManager(&commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: *commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	chartDir, err := common.GetHelmChartDir(werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %s", err)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	buildOptions, err := common.GetBuildOptions(&commonCmdData, werfConfig)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	repoAddress, err := common.GetStagesStorageAddress(&commonCmdData)
	if err != nil {
		return err
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepository string

	if len(werfConfig.StapelImages) != 0 || len(werfConfig.ImagesFromDockerfile) != 0 {
		containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO
		stagesStorage, err := common.GetStagesStorage(repoAddress, containerRuntime, &commonCmdData)
		if err != nil {
			return err
		}
		synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
		if err != nil {
			return err
		}
		stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
		if err != nil {
			return err
		}
		storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
		if err != nil {
			return err
		}
		secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerRuntime, &commonCmdData)
		if err != nil {
			return err
		}

		storageManager := manager.NewStorageManager(projectName, stagesStorage, secondaryStagesStorageList, storageLockManager, stagesStorageCache)

		imagesRepository = storageManager.StagesStorage.String()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(&commonCmdData, buildOptions)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, nil, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, storageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if *commonCmdData.SkipBuild {
				if err := c.ShouldBeBuilt(ctx); err != nil {
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
	}

	secretsManager := secrets_manager.NewSecretsManager(giterminismManager.ProjectDir(), secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	wc := chart_extender.NewWerfChart(ctx, giterminismManager, secretsManager, chartDir, cmd_helm.Settings, chart_extender.WerfChartOptions{
		SecretValueFiles: *commonCmdData.SecretValues,
		ExtraAnnotations: userExtraAnnotations,
		ExtraLabels:      userExtraLabels,
	})

	if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
		return err
	}
	if err := wc.SetWerfConfig(werfConfig); err != nil {
		return err
	}
	if vals, err := chart_extender.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, imagesInfoGetters, chart_extender.ServiceValuesOptions{Env: *commonCmdData.Environment}); err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	} else if err := wc.SetServiceValues(vals); err != nil {
		return err
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, nil, "", cmd_helm.Settings, actionConfig, helm.InitActionConfigOptions{}); err != nil {
		return err
	}

	cmd_helm.Settings.Debug = *commonCmdData.LogDebug

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender:               wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender { return chart_extender.NewWerfSubchart() },
	}

	// FIXME! do not run render during publish, only save passed values into the bundle
	valueOpts := &values.Options{
		ValueFiles:   *commonCmdData.Values,
		StringValues: *commonCmdData.SetString,
		Values:       *commonCmdData.Set,
		FileValues:   *commonCmdData.SetFile,
	}

	postRenderer, err := wc.GetPostRenderer()
	if err != nil {
		return err
	}

	helmTemplateCmd, _ := cmd_helm.NewTemplateCmd(actionConfig, ioutil.Discard, cmd_helm.TemplateCmdOptions{
		PostRenderer: postRenderer,
		ValueOpts:    valueOpts,
	})
	if err := helmTemplateCmd.RunE(helmTemplateCmd, []string{"RELEASE", filepath.Join(giterminismManager.ProjectDir(), chartDir)}); err != nil {
		return err
	}

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewV4().String())
	defer os.RemoveAll(bundleTmpDir)

	p := getter.All(cmd_helm.Settings)
	if vals, err := valueOpts.MergeValues(p, loader.GlobalLoadOptions.ChartExtender); err != nil {
		return err
	} else if bundle, err := wc.CreateNewBundle(ctx, bundleTmpDir, vals); err != nil {
		return fmt.Errorf("unable to create bundle: %s", err)
	} else {
		loader.GlobalLoadOptions = &loader.LoadOptions{}

		bundleRef := fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag)

		if err := logboek.Context(ctx).LogProcess("Saving bundle to the local chart helm cache").DoError(func() error {
			helmChartSaveCmd := cmd_helm.NewChartSaveCmd(actionConfig, logboek.Context(ctx).OutStream())
			if err := helmChartSaveCmd.RunE(helmChartSaveCmd, []string{bundle.Dir, bundleRef}); err != nil {
				return fmt.Errorf("error saving bundle to the local chart helm cache: %s", err)
			}
			return nil
		}); err != nil {
			return err
		}

		if err := logboek.Context(ctx).LogProcess("Pushing bundle %q", bundleRef).DoError(func() error {
			helmChartPushCmd := cmd_helm.NewChartPushCmd(actionConfig, logboek.Context(ctx).OutStream())
			if err := helmChartPushCmd.RunE(helmChartPushCmd, []string{bundleRef}); err != nil {
				return fmt.Errorf("error pushing bundle %q: %s", bundleRef, err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
