package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge",
		DisableFlagsInUseLine: true,
		Short:                 "Purge all project images from images repo and stages from stages storage",
		Long: common.GetLongCommandDescription(`Purge all project images from images repo and stages from stages storage.

First step is 'werf images purge', which will delete all project images from images repo. Second step is 'werf stages purge', which will delete all stages from stages storage.

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually. Images from images repo, that are being used in Kubernetes cluster will also be deleted.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ApplyLogOptions(&CommonCmdData); err != nil {
				cmd.Help()
				fmt.Println()
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runPurge()
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupStagesStorage(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to delete images from the specified stages storage and images repo.")
	common.SetupInsecureRepo(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)

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

	_, err = common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{AllowInsecureRepo: *CommonCmdData.InsecureRepo}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	var imageNames []string
	for _, image := range werfConfig.Images {
		imageNames = append(imageNames, image.Name)
	}

	commonRepoOptions := cleaning.CommonRepoOptions{
		ImagesRepo:  imagesRepo,
		ImagesNames: imageNames,
		DryRun:      *CommonCmdData.DryRun,
	}

	commonProjectOptions := cleaning.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleaning.CommonOptions{DryRun: *CommonCmdData.DryRun},
	}

	purgeOptions := cleaning.PurgeOptions{
		CommonRepoOptions:    commonRepoOptions,
		CommonProjectOptions: commonProjectOptions,
	}

	logger.OptionalLnModeOn()
	if err := cleaning.Purge(purgeOptions); err != nil {
		return err
	}

	return nil
}
