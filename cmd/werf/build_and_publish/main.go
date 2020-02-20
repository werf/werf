package build_and_publish

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	PullUsername string
	PullPassword string

	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build-and-publish [IMAGE_NAME...]",
		Short: "Build stages and publish images",
		Long: common.GetLongCommandDescription(`Build stages and final images using each specified tag with the tagging strategy and push into images repo.

Command combines 'werf stages build' and 'werf images publish'.

After stages has been built, new docker layer with service info about tagging strategy will be built for each tag of each image from werf.yaml. Images will be pushed into docker repo with the names IMAGES_REPO/IMAGE_NAME:TAG. See more info about publish process: https://werf.io/documentation/reference/publish_process.html.

The result of build-and-publish command is stages in stages storage and named images pushed into the docker repo.

If one or more IMAGE_NAME parameters specified, werf will build images stages and publish only these images from werf.yaml`),
		Example: `  # Build and publish all images from werf.yaml into specified docker repo, built stages will be placed locally; tag images with the mytag tag using custom tagging strategy
  $ werf build-and-publish --stages-storage :local --images-repo registry.mydomain.com/myproject --tag-custom mytag

  # Build and publish all images from werf.yaml into minikube registry; tag images with the mybranch tag, using git-branch tagging strategy
  $ werf build-and-publish --stages-storage :local --images-repo :minikube --tag-git-branch mybranch

  # Build and publish with enabled drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf build-and-publish --stages-storage :local --introspect-error --images-repo :minikube --tag-git-branch mybranch

  # Set --stages-storage default value using $WERF_STAGES_STORAGE param and --images-repo default value using $WERF_IMAGE_REPO param
  $ export WERF_STAGES_STORAGE=:local
  $ export WERF_IMAGES_REPO=myregistry.mydomain.com/myproject
  $ werf build-and-publish --tag-git-tag v1.4.9`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runBuildAndPublish(args)
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	common.SetupTag(&CommonCmdData, cmd)
	common.SetupStagesStorage(&CommonCmdData, cmd)
	common.SetupStagesStorageLock(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepoMode(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified stages storage, to push images into the specified images repo, to pull base images")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)
	common.SetupLogProjectDir(&CommonCmdData, cmd)

	common.SetupIntrospectStage(&CommonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
	cmd.Flags().BoolVarP(&CmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")

	return cmd
}

func runBuildAndPublish(imagesToProcess []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *CommonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *CommonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&CommonCmdData, projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	for _, imageToProcess := range imagesToProcess {
		if !werfConfig.HasImage(imageToProcess) {
			return fmt.Errorf("specified image %s is not defined in werf.yaml", logging.ImageLogName(imageToProcess, false))
		}
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoMode, err := common.GetImagesRepoMode(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoManager, err := common.GetImagesRepoManager(imagesRepo, imagesRepoMode)
	if err != nil {
		return err
	}

	stagesStorage, err := common.GetStagesStorage(&CommonCmdData)
	if err != nil {
		return err
	}

	tagOpts, err := common.GetTagOptions(&CommonCmdData, common.TagOptionsGetterOptions{})
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogWarnF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	introspectOptions, err := common.GetIntrospectOptions(&CommonCmdData, werfConfig)
	if err != nil {
		return err
	}

	opts := build.BuildAndPublishOptions{
		BuildStagesOptions: build.BuildStagesOptions{
			ImageBuildOptions: image.BuildOptions{
				IntrospectAfterError:  CmdData.IntrospectAfterError,
				IntrospectBeforeError: CmdData.IntrospectBeforeError,
			},
			IntrospectOptions: introspectOptions,
		},
		PublishImagesOptions: build.PublishImagesOptions{
			ImagesToPublish: imagesToProcess,
			TagOptions:      tagOpts,
		},
	}

	c := build.NewConveyor(werfConfig, imagesToProcess, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)
	defer c.Terminate()

	if err = c.BuildAndPublish(stagesStorage, imagesRepoManager, opts); err != nil {
		return err
	}

	return nil
}
