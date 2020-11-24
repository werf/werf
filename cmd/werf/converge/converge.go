package converge

import (
	"context"
	"fmt"
	"time"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/deploy/werf_chart"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "converge",
		Short: "Build and push images, then deploy application into Kubernetes",
		Long: common.GetLongCommandDescription(`Build and push images, then deploy application into Kubernetes.

The result of converge command is an application deployed into Kubernetes for current git state. Command will create release and wait until all resources of the release will become ready.

Environment is a required param for the deploy by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/documentation/advanced/helm/basics.html`),
		Example: `# Build and deploy current application state into production environment
werf converge --repo registry.mydomain.com/web --env production`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.BackgroundContext()
			defer werf.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runMain(ctx)
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupDisableDeterminism(&commonCmdData, cmd)
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

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
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

	common.SetupFollow(&commonCmdData, cmd)

	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "auto-rollback", "R", common.GetBoolEnvironmentDefaultFalse("WERF_AUTO_ROLLBACK"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "atomic", "", common.GetBoolEnvironmentDefaultFalse("WERF_ATOMIC"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_ATOMIC by default)")

	return cmd
}

func runMain(ctx context.Context) error {
	tmp_manager.AutoGCEnabled = true

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

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	if err := ssh_agent.Init(ctx, *commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	if err := kube.Init(kube.InitOptions{kube.KubeConfigOptions{
		Context:          *commonCmdData.KubeContext,
		ConfigPath:       *commonCmdData.KubeConfig,
		ConfigDataBase64: *commonCmdData.KubeConfigBase64,
	}}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(ctx); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	if *commonCmdData.Follow {
		logboek.LogOptionalLn()
		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context) error {
			return run(ctx, projectDir)
		})
	} else {
		return run(ctx, projectDir)
	}
}

func run(ctx context.Context, projectDir string) error {
	localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
	if err != nil {
		return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, localGitRepo, config.WerfConfigOptions{LogRenderedFilePath: true, DisableDeterminism: *commonCmdData.DisableDeterminism, Env: *commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	chartDir, err := common.GetHelmChartDir(projectDir, &commonCmdData, werfConfig)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	buildOptions, err := common.GetBuildOptions(&commonCmdData, werfConfig)
	if err != nil {
		return err
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepository string
	if len(werfConfig.StapelImages) != 0 || len(werfConfig.ImagesFromDockerfile) != 0 {
		stagesStorageAddress, err := common.GetStagesStorageAddress(&commonCmdData)
		if err != nil {
			return err
		}
		containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO
		stagesStorage, err := common.GetStagesStorage(stagesStorageAddress, containerRuntime, &commonCmdData)
		if err != nil {
			return err
		}
		logboek.LogOptionalLn()
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

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, localGitRepo, nil, projectDir, projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, storageLockManager, conveyorOptions)
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

	var secretsManager secret.Manager
	if m, err := deploy.GetSafeSecretManager(ctx, projectDir, chartDir, *commonCmdData.SecretValues, localGitRepo, *commonCmdData.DisableDeterminism, *commonCmdData.IgnoreSecretKey); err != nil {
		return err
	} else {
		secretsManager = m
	}

	releaseName, err := common.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*commonCmdData.Namespace, *commonCmdData.Environment, werfConfig)
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

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %s", err)
	} else {
		lockManager = m
	}

	wc := werf_chart.NewWerfChart(ctx, localGitRepo, *commonCmdData.DisableDeterminism, projectDir, werf_chart.WerfChartOptions{
		ReleaseName: releaseName,
		ChartDir:    chartDir,

		SecretValueFiles: *commonCmdData.SecretValues,
		ExtraAnnotations: userExtraAnnotations,
		ExtraLabels:      userExtraLabels,

		LockManager:    lockManager,
		SecretsManager: secretsManager,
	})
	if err := wc.SetEnv(*commonCmdData.Environment); err != nil {
		return err
	}
	if err := wc.SetWerfConfig(werfConfig); err != nil {
		return err
	}
	if vals, err := werf_chart.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepository, namespace, imagesInfoGetters, werf_chart.ServiceValuesOptions{Env: *commonCmdData.Environment}); err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	} else if err := wc.SetServiceValues(vals); err != nil {
		return err
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, namespace, cmd_helm.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
	}); err != nil {
		return err
	}

	helmUpgradeCmd, _ := cmd_helm.NewUpgradeCmd(actionConfig, logboek.ProxyOutStream(), cmd_helm.UpgradeCmdOptions{
		LoadOptions: loader.LoadOptions{
			ChartExtender: wc,
			SubchartExtenderFactoryFunc: func() chart.ChartExtender {
				return werf_chart.NewWerfChart(ctx, nil, *commonCmdData.DisableDeterminism, projectDir, werf_chart.WerfChartOptions{})
			},
			LoadDirFunc:     common.MakeChartDirLoadFunc(ctx, localGitRepo, projectDir, *commonCmdData.DisableDeterminism),
			LocateChartFunc: common.MakeLocateChartFunc(ctx, localGitRepo, projectDir, *commonCmdData.DisableDeterminism),
			ReadFileFunc:    common.MakeHelmReadFileFunc(ctx, localGitRepo, projectDir, *commonCmdData.DisableDeterminism),
		},
		PostRenderer: wc.ExtraAnnotationsAndLabelsPostRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   *commonCmdData.Values,
			StringValues: *commonCmdData.SetString,
			Values:       *commonCmdData.Set,
			FileValues:   *commonCmdData.SetFile,
		},
		CreateNamespace: NewBool(true),
		Install:         NewBool(true),
		Wait:            NewBool(true),
		Atomic:          NewBool(cmdData.AutoRollback),
		Timeout:         NewDuration(time.Duration(cmdData.Timeout)),
	})
	return wc.WrapUpgrade(ctx, func() error {
		return helmUpgradeCmd.RunE(helmUpgradeCmd, []string{releaseName, chartDir})
	})
}

func NewDuration(value time.Duration) *time.Duration {
	if value != 0 {
		res := new(time.Duration)
		*res = value
		return res
	}
	return nil
}

func NewBool(value bool) *bool {
	res := new(bool)
	*res = value
	return res
}
