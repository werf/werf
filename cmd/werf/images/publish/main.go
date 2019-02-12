package publish

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

type CmdDataType struct {
}

var CmdData CmdDataType
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	return NewCmdWithData(&CommonCmdData)
}

func NewCmdWithData(commonCmdData *common.CmdData) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish [IMAGE_NAME...]",
		Short: "Build images and push into images repo.",
		Long: common.GetLongCommandDescription(`Build final images using each specified tag with the tagging strategy and push into images repo.

New docker layer with service info about tagging strategy will be built for each tag of each image from werf.yaml. Images will be pushed into docker repo with the names IMAGES_REPO/IMAGE_NAME:TAG. See more info about images naming: https://flant.github.io/werf/reference/registry/image_naming.html.

If one or more IMAGE_NAME parameters specified, werf will publish only these images from werf.yaml.`),
		Example: `  # Publish images into myregistry.mydomain.com/myproject images repo using 'mybranch' tag and git-branch tagging strategy
  $ werf images publish --stages-storage :local --images-repo myregistry.mydomain.com/myproject --tag-git-branch mybranch`,
		DisableFlagsInUseLine: true,
		Annotations:           map[string]string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.LogRunningTime(func() error {
				common.LogVersion()

				return runImagesPublish(commonCmdData, args)
			})
		},
	}

	common.SetupDir(commonCmdData, cmd)
	common.SetupTmpDir(commonCmdData, cmd)
	common.SetupHomeDir(commonCmdData, cmd)
	common.SetupSSHKey(commonCmdData, cmd)

	common.SetupTag(commonCmdData, cmd)

	common.SetupStagesRepo(commonCmdData, cmd)
	common.SetupImagesRepo(commonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage and push images into images repo.")

	return cmd
}

func runImagesPublish(commonCmdData *common.CmdData, imagesToProcess []string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logger.GetOutStream(), Err: logger.GetErrStream()}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	_, err = common.GetStagesRepo(commonCmdData)
	if err != nil {
		return err
	}

	imagesRepo, err := common.GetImagesRepo(projectName, commonCmdData)
	if err != nil {
		return err
	}

	tagOpts, err := common.GetTagOptions(commonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogErrorF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	opts := build.PublishImagesOptions{TagOptions: tagOpts}

	c := build.NewConveyor(werfConfig, imagesToProcess, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)

	if err = c.PublishImages(imagesRepo, opts); err != nil {
		return err
	}

	return nil
}
