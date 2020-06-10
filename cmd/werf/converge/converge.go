package converge

import (
	"fmt"
	"time"

	"github.com/werf/werf/pkg/tag_strategy"

	"github.com/werf/werf/pkg/images_manager"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"

	"github.com/werf/werf/pkg/image"

	"github.com/werf/werf/pkg/stages_manager"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	PullUsername string
	PullPassword string
	Timeout      int
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "converge",
		Short: "Build stages and publish images, then deploy application into Kubernetes",
		Long: common.GetLongCommandDescription(`Build stages and final images using content based tagging and publish into images repo, then deploy application chart.

Command combines 'werf stages build', 'werf images publish' and 'werf deploy'.

The result of converge command is an application deployed into Kubernetes for current git state. Command will create release and wait until all resources of the release will become ready.

Environment is a required param for the deploy by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html`),
		Example: `# Build and deploy current application state into production environment
werf converge --stages-storage registry.mydomain.com/web/back/stages --images-repo registry.mydomain.com/web/back --env production`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs, common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runConverge()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupImagesRepoOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified stages storage, to push images into the specified images repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmChartDir(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(&commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(&commonCmdData, cmd)

	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	return cmd
}

func runConverge() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(&commonCmdData, stagesStorage.Address())
	if err != nil {
		return err
	}
	stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(synchronization)
	if err != nil {
		return err
	}

	stagesManager := stages_manager.NewStagesManager(projectName, storageLockManager, stagesStorageCache)
	if err := stagesManager.UseStagesStorage(stagesStorage); err != nil {
		return err
	}

	imagesRepo, err := common.GetImagesRepo(projectName, &commonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogWarnF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*commonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	helmChartDir, err := common.GetHelmChartDir(projectDir, &commonCmdData)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %s", err)
	}

	release, err := common.GetHelmRelease(*commonCmdData.Release, *commonCmdData.Environment, werfConfig)
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

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *commonCmdData.KubeConfig,
			KubeContext:                 *commonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *commonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			StatusProgressPeriod:        common.GetStatusProgressPeriod(&commonCmdData),
			HooksStatusProgressPeriod:   common.GetHooksStatusProgressPeriod(&commonCmdData),
			ReleasesMaxHistory:          *commonCmdData.ReleasesHistoryMax,
			InitNamespace:               true,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	opts := build.BuildAndPublishOptions{
		BuildStagesOptions: build.BuildStagesOptions{
			ImageBuildOptions: container_runtime.BuildOptions{},
		},
		PublishImagesOptions: build.PublishImagesOptions{
			TagOptions: build.TagOptions{TagByStagesSignature: true}, // always content based tagging
		},
	}

	logboek.LogOptionalLn()

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, nil, projectDir, projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, stagesManager, imagesRepo, storageLockManager, common.GetConveyorOptions(&commonCmdData))
	defer conveyorWithRetry.Terminate()

	var imagesInfoGetters []images_manager.ImageInfoGetter

	if err := conveyorWithRetry.WithRetryBlock(func(c *build.Conveyor) error {
		if err := c.BuildAndPublish(opts); err != nil {
			return err
		}

		imagesInfoGetters = c.GetImageInfoGetters(werfConfig.StapelImages, werfConfig.ImagesFromDockerfile, "", tag_strategy.StagesSignature, false)

		return nil
	}); err != nil {
		return err
	}

	logboek.LogOptionalLn()
	return deploy.Deploy(projectName, projectDir, helmChartDir, imagesRepo.String(), imagesInfoGetters, release, namespace, "", tag_strategy.StagesSignature, werfConfig, *commonCmdData.HelmReleaseStorageNamespace, helmReleaseStorageType, storageLockManager, deploy.DeployOptions{
		Set:                  *commonCmdData.Set,
		SetString:            *commonCmdData.SetString,
		Values:               *commonCmdData.Values,
		SecretValues:         *commonCmdData.SecretValues,
		Timeout:              time.Duration(cmdData.Timeout) * time.Second,
		Env:                  *commonCmdData.Environment,
		UserExtraAnnotations: userExtraAnnotations,
		UserExtraLabels:      userExtraLabels,
		IgnoreSecretKey:      *commonCmdData.IgnoreSecretKey,
		ThreeWayMergeMode:    helm.ThreeWayMergeEnabled,
	})

	return nil
}
