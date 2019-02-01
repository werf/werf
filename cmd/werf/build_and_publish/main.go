package build_and_publish

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/ssh_agent"
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
		Short: "Build stages then publish images into Docker registry",
		Long: common.GetLongCommandDescription(`Build stages then publish images into Docker registry.

New docker layer with service info about tagging scheme will be built for each image. Images will be pushed into docker registry with the names IMAGE_REPO/IMAGE_NAME:TAG. See more info about images naming: https://flant.github.io/werf/reference/registry/image_naming.html.

The result of build-and-publish command is a stages cache for images and named images pushed into the docker registry.

If one or more IMAGE_NAME parameters specified, werf will build images stages and publish only these images from werf.yaml.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfAnsibleArgs, common.WerfDockerConfig, common.WerfHome, common.WerfTmp),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			if CmdData.PullUsername == "" {
				CmdData.PullUsername = *CommonCmdData.ImagesUsername
			}
			if CmdData.PullPassword == "" {
				CmdData.PullPassword = *CommonCmdData.ImagesPassword
			}

			return common.LogRunningTime(func() error {
				err := runBuildAndPublish(args)
				if err != nil {
					return fmt.Errorf("build-and-publish failed: %s", err)
				}

				return nil
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.Flags().StringVarP(&CmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")

	cmd.Flags().BoolVarP(&CmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
	cmd.Flags().BoolVarP(&CmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")

	common.SetupTag(&CommonCmdData, cmd)
	common.SetupStagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesUsername(&CommonCmdData, cmd, "Docker registry username to authorize pull of base images and push to the docker imagesRepo")
	common.SetupImagesPassword(&CommonCmdData, cmd, "Docker registry password to authorize pull of base images and push to the docker imagesRepo")

	return cmd
}

func runBuildAndPublish(imagesToProcess []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetBuildAndPublishDockerAuthorizer(projectTmpDir, CmdData.PullUsername, CmdData.PullPassword, *CommonCmdData.ImagesUsername, *CommonCmdData.ImagesPassword)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogWarningF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	stagesRepo, err := common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	tagOpts, err := common.GetTagOptions(&CommonCmdData)
	if err != nil {
		return err
	}

	opts := build.BuildAndPublishOptions{
		BuildStagesOptions: build.BuildStagesOptions{
			ImageBuildOptions: image.BuildOptions{
				IntrospectAfterError:  CmdData.IntrospectAfterError,
				IntrospectBeforeError: CmdData.IntrospectBeforeError,
			},
		},
		PublishImagesOptions: build.PublishImagesOptions{
			TagOptions: tagOpts,
		},
	}

	c := build.NewConveyor(werfConfig, imagesToProcess, projectDir, common.GetSharedContextDir(), common.GetLocalCacheDir(), projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)

	if err = c.BuildAndPublish(stagesRepo, imagesRepo, opts); err != nil {
		return err
	}

	return nil
}
