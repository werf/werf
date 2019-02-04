package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "purge",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			err := runPurge()
			if err != nil {
				return fmt.Errorf("images purge failed: %s", err)
			}

			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesUsernameWithUsage(&CommonCmdData, cmd, "Images Docker repo username (granted permission to read images info and delete images)")
	common.SetupImagesPasswordWithUsage(&CommonCmdData, cmd, "Images Docker repo username (granted permission to read images info and delete images)")

	cmd.Flags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runPurge() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
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

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	dockerAuthorizer, err := docker_authorizer.GetDockerAuthorizer(projectTmpDir, *CommonCmdData.ImagesUsername, *CommonCmdData.ImagesPassword)
	if err != nil {
		return err
	}

	if err := dockerAuthorizer.Login(imagesRepo); err != nil {
		return err
	}

	var imageNames []string
	for _, image := range werfConfig.Images {
		imageNames = append(imageNames, image.Name)
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		ImagesRepo:  imagesRepo,
		ImagesNames: imageNames,
		DryRun:      CmdData.DryRun,
	}

	if err := cleanup.ImagesPurge(commonRepoOptions); err != nil {
		return err
	}

	return nil
}
