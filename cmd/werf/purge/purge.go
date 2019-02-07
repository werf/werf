package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "purge",
		DisableFlagsInUseLine: true,
		Short: "Purge all project images from images repo and stages from stages repo (or locally)",
		Long: common.GetLongCommandDescription(`Purge all project images from images repo and stages from stages repo (or locally).

First step is 'werf images purge', which will delete all project images from images repo. Second step is 'werf stages purge', which will delete all stages from stages repo (or locally).

Command allows deletion of all images of the project at once, meant to be used manually.

WARNING: Images from images repo, that are being used in Kubernetes cluster will also be deleted.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			return runPurge()
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupStagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

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

	_, err = common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	var imageNames []string
	for _, image := range werfConfig.Images {
		imageNames = append(imageNames, image.Name)
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		ImagesRepo:  imagesRepo,
		ImagesNames: imageNames,
		DryRun:      CommonCmdData.DryRun,
	}

	if err := cleanup.ImagesPurge(commonRepoOptions); err != nil {
		return err
	}

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: CommonCmdData.DryRun},
	}

	if err := cleanup.StagesPurge(commonProjectOptions); err != nil {
		return err
	}

	return nil
}
