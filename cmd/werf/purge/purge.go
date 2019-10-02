package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Force bool
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
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
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
	common.SetupImagesRepoMode(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to delete images from the specified stages storage and images repo")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)
	common.SetupLogProjectDir(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)
	cmd.Flags().BoolVarP(&CmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

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

	common.ProcessLogProjectDir(&CommonCmdData, projectDir)

	_, err = common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *CommonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *CommonCmdData.SkipTlsVerifyRegistry}); err != nil {
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

	imagesRepoMode, err := common.GetImagesRepoMode(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoManager, err := common.GetImagesRepoManager(imagesRepo, imagesRepoMode)
	if err != nil {
		return err
	}

	var imageNames []string
	for _, image := range werfConfig.StapelImages {
		imageNames = append(imageNames, image.Name)
	}

	for _, image := range werfConfig.ImagesFromDockerfile {
		imageNames = append(imageNames, image.Name)
	}

	imagesPurgeOptions := cleaning.ImagesPurgeOptions{
		ImagesRepoManager: imagesRepoManager,
		ImagesNames:       imageNames,
		DryRun:            *CommonCmdData.DryRun,
	}

	stagesPurgeOptions := cleaning.StagesPurgeOptions{
		ProjectName:                   projectName,
		RmContainersThatUseWerfImages: CmdData.Force,
		DryRun:                        *CommonCmdData.DryRun,
	}

	purgeOptions := cleaning.PurgeOptions{
		ImagesPurgeOptions: imagesPurgeOptions,
		StagesPurgeOptions: stagesPurgeOptions,
	}

	logboek.LogOptionalLn()
	if err := cleaning.Purge(purgeOptions); err != nil {
		return err
	}

	return nil
}
