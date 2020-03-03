package deploy

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/flant/werf/pkg/images_manager"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var cmdData struct {
	Timeout int
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application into Kubernetes",
		Long: common.GetLongCommandDescription(`Deploy application into Kubernetes.

Command will create Helm Release and wait until all resources of the release are become ready.

Deploy needs the same parameters as push to construct image names: repo and tags. Docker images names are constructed from parameters as IMAGES_REPO/IMAGE_NAME:TAG. Deploy will fetch built image ids from Docker repo. So images should be published prior running deploy.

Helm chart directory .helm should exists and contain valid Helm chart.

Environment is a required param for the deploy by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or $WERF_ENV should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://werf.io/documentation/reference/deploy_process/deploy_into_kubernetes.html`),
		Example: `  # Deploy project named 'myproject' into 'dev' environment using images from registry.mydomain.com/myproject tagged as mytag with git-tag tagging strategy; helm release name and namespace will be named as 'myproject-dev'
  $ werf deploy --stages-storage :local --env dev --images-repo registry.mydomain.com/myproject --tag-git-tag mytag

  # Deploy project using specified helm release name and namespace using images from registry.mydomain.com/myproject tagged with docker tag 'myversion'
  $ werf deploy --stages-storage :local --release myrelease --namespace myns --images-repo registry.mydomain.com/myproject --tag-custom myversion`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runDeploy()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupTag(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)
	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupStagesStorage(&commonCmdData, cmd)
	common.SetupStagesStorageLock(&commonCmdData, cmd)
	common.SetupImagesRepo(&commonCmdData, cmd)
	common.SetupImagesRepoMode(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage and images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	common.SetupThreeWayMergeMode(&commonCmdData, cmd)

	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	return cmd
}

func runDeploy() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*commonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	threeWayMergeMode, err := common.GetThreeWayMergeMode(*commonCmdData.ThreeWayMergeMode)
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

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *commonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	var imagesRepoManager *common.ImagesRepoManager
	var tag string
	var tagStrategy tag_strategy.TagStrategy
	var imagesInfoGetters []images_manager.ImageInfoGetter
	if len(werfConfig.StapelImages) != 0 || len(werfConfig.ImagesFromDockerfile) != 0 {
		if len(werfConfig.StapelImages) != 0 {
			_, err = common.GetStagesStorage(&commonCmdData)
			if err != nil {
				return err
			}

			_, err = common.GetStagesStorageLock(&commonCmdData)
			if err != nil {
				return err
			}
		}

		imagesRepo, err := common.GetImagesRepo(werfConfig.Meta.Project, &commonCmdData)
		if err != nil {
			return err
		}

		imagesRepoMode, err := common.GetImagesRepoMode(&commonCmdData)
		if err != nil {
			return err
		}

		imagesRepoManager, err = common.GetImagesRepoManager(imagesRepo, imagesRepoMode)
		if err != nil {
			return err
		}

		tag, tagStrategy, err = common.GetDeployTag(&commonCmdData, common.TagOptionsGetterOptions{})
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

		logboek.LogOptionalLn()
		c := build.NewConveyor(werfConfig, []string{}, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)
		defer c.Terminate()

		if err = c.ShouldBeBuilt(); err != nil {
			return err
		}

		imagesInfoGetters = c.GetImageInfoGetters(werfConfig.StapelImages, werfConfig.ImagesFromDockerfile, imagesRepoManager, tag, tagStrategy, false)
	}

	if imagesRepoManager == nil {
		imagesRepoManager = &common.ImagesRepoManager{}
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

	logboek.LogOptionalLn()
	return deploy.Deploy(projectDir, imagesRepoManager, imagesInfoGetters, release, namespace, tag, tagStrategy, werfConfig, *commonCmdData.HelmReleaseStorageNamespace, helmReleaseStorageType, deploy.DeployOptions{
		Set:                  *commonCmdData.Set,
		SetString:            *commonCmdData.SetString,
		Values:               *commonCmdData.Values,
		SecretValues:         *commonCmdData.SecretValues,
		Timeout:              time.Duration(cmdData.Timeout) * time.Second,
		Env:                  *commonCmdData.Environment,
		UserExtraAnnotations: userExtraAnnotations,
		UserExtraLabels:      userExtraLabels,
		IgnoreSecretKey:      *commonCmdData.IgnoreSecretKey,
		ThreeWayMergeMode:    threeWayMergeMode,
	})
}
