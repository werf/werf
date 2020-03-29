package mv

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/werf/pkg/image"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/storage"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var cmdData struct {
	FromStagesStorage         *string
	FromStagesStorageRepoData *common.RepoData

	ToStagesStorage         *string
	ToStagesStorageRepoData *common.RepoData
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "mv",
		DisableFlagsInUseLine: true,
		Short:                 "Move project stages from one stages storage to another",
		Long:                  common.GetLongCommandDescription("Move project stages from one stages storage to another"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runMv()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storages")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	SetupFromStagesStorage(cmd)
	SetupToStagesStorage(cmd)

	return cmd
}

func SetupFromStagesStorage(cmd *cobra.Command) {
	if commonCmdData.CommonRepoData == nil {
		common.SetupCommonRepoData(&commonCmdData, cmd)
	}

	cmdData.FromStagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.FromStagesStorage, "from-stages-storage", "", os.Getenv("WERF_FROM_STAGES_STORAGE"), fmt.Sprintf("Source stages storage from which stages will be moved (docker repo address or :local should be specified, default $WERF_FROM_STAGES_STORAGE environment)."))

	cmdData.FromStagesStorageRepoData = &common.RepoData{DesignationStorageName: "source stages storage"}

	common.SetupImplementationForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-implementation", []string{"WERF_FROM_STAGES_STORAGE_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	common.SetupDockerHubUsernameForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-docker-hub-username", []string{"WERF_FROM_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	common.SetupDockerHubPasswordForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-docker-hub-password", []string{"WERF_FROM_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	common.SetupGithubTokenForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-github-token", []string{"WERF_FROM_STAGES_STORAGE_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	common.SetupHarborUsernameForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-harbor-username", []string{"WERF_FROM_STAGES_STORAGE_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	common.SetupHarborPasswordForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-stages-storage-repo-harbor-password", []string{"WERF_FROM_STAGES_STORAGE_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
}

func SetupToStagesStorage(cmd *cobra.Command) {
	if commonCmdData.CommonRepoData == nil {
		common.SetupCommonRepoData(&commonCmdData, cmd)
	}

	cmdData.ToStagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.ToStagesStorage, "to-stages-storage", "", os.Getenv("WERF_TO_STAGES_STORAGE"), fmt.Sprintf("Destination stages storage to which stages will be moved (docker repo address or :local should be specified, default $WERF_TO_STAGES_STORAGE environment)."))

	cmdData.ToStagesStorageRepoData = &common.RepoData{DesignationStorageName: "destination stages storage"}

	common.SetupImplementationForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-implementation", []string{"WERF_TO_STAGES_STORAGE_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	common.SetupDockerHubUsernameForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-docker-hub-username", []string{"WERF_TO_STAGES_STORAGE_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	common.SetupDockerHubPasswordForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-docker-hub-password", []string{"WERF_TO_STAGES_STORAGE_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	common.SetupGithubTokenForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-github-token", []string{"WERF_TO_STAGES_STORAGE_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	common.SetupHarborUsernameForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-harbor-username", []string{"WERF_TO_STAGES_STORAGE_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	common.SetupHarborPasswordForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-stages-storage-repo-harbor-password", []string{"WERF_TO_STAGES_STORAGE_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
}

func NewFromStagesStorage(containerRuntime container_runtime.ContainerRuntime) (storage.StagesStorage, error) {
	if *cmdData.FromStagesStorage == "" {
		return nil, fmt.Errorf("--from-stages-storage=ADDRESS param required")
	}

	repoData := common.MergeRepoData(cmdData.FromStagesStorageRepoData, commonCmdData.CommonRepoData)

	if err := common.ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return NewStagesStorage(*cmdData.FromStagesStorage, repoData, containerRuntime)
}

func NewToStagesStorage(containerRuntime container_runtime.ContainerRuntime) (storage.StagesStorage, error) {
	if *cmdData.ToStagesStorage == "" {
		return nil, fmt.Errorf("--to-stages-storage=ADDRESS param required")
	}

	repoData := common.MergeRepoData(cmdData.ToStagesStorageRepoData, commonCmdData.CommonRepoData)

	if err := common.ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return NewStagesStorage(*cmdData.ToStagesStorage, repoData, containerRuntime)
}

func NewStagesStorage(stagesStorageAddress string, repoData *common.RepoData, containerRuntime container_runtime.ContainerRuntime) (storage.StagesStorage, error) {
	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerRuntime,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				Implementation: *repoData.Implementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *commonCmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *repoData.DockerHubUsername,
					DockerHubPassword:     *repoData.DockerHubPassword,
					DockerHubToken:        *repoData.DockerHubToken,
					GitHubToken:           *repoData.GitHubToken,
					HarborUsername:        *repoData.HarborUsername,
					HarborPassword:        *repoData.HarborPassword,
				},
			},
		},
	)
}

func runMv() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
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

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	logboek.LogOptionalLn()

	projectName := werfConfig.Meta.Project

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	fromStagesStorage, err := NewFromStagesStorage(containerRuntime)
	if err != nil {
		return err
	}

	toStagesStorage, err := NewToStagesStorage(containerRuntime)
	if err != nil {
		return err
	}

	_, err = common.GetSynchronization(&commonCmdData)
	if err != nil {
		return err
	}

	return SyncStages(projectName, fromStagesStorage, toStagesStorage, containerRuntime)
}

// SyncStages will make sure, that destination stages storage contains all stages from source stages storage.
// Repeatedly calling SyncStages will copy stages from source stages storage to destination, that already exists in the destination.
// SyncStages will not delete excess stages from destination storage, that does not exists in the source.
func SyncStages(projectName string, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime) error {
	isOk := false
	logboek.Default.LogProcessStart(fmt.Sprintf("Sync %q project stages", projectName), logboek.LevelLogProcessStartOptions{})
	defer func() {
		if isOk {
			logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
		} else {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		}
	}()

	logboek.Default.LogFDetails("Source      — %s\n", fromStagesStorage.String())
	logboek.Default.LogFDetails("Destination — %s\n", toStagesStorage.String())
	logboek.Default.LogOptionalLn()

	var errors []error

	getAllRepoImagesFunc := func(logProcessMsg string, stagesStorage storage.StagesStorage) ([]*image.Info, error) {
		logboek.Default.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
		if repoImages, err := stagesStorage.GetRepoImages(projectName); err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return nil, fmt.Errorf("unable to get repo images from %s: %s", fromStagesStorage.String(), err)
		} else {
			logboek.Default.LogFDetails("Stages count: %d\n", len(repoImages))
			logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
			return repoImages, nil
		}
	}

	var existingSourceRepoImages []*image.Info
	var existingDestinationRepoImages []*image.Info

	if repoImages, err := getAllRepoImagesFunc("Getting all repo images list from source stages storage", fromStagesStorage); err != nil {
		return fmt.Errorf("unable to get repo images from source %s: %s", fromStagesStorage.String(), err)
	} else {
		existingSourceRepoImages = repoImages
	}

	if repoImages, err := getAllRepoImagesFunc("Getting all repo images list from destination stages storage", toStagesStorage); err != nil {
		return fmt.Errorf("unable to get repo images from destination %s: %s", toStagesStorage.String(), err)
	} else {
		existingDestinationRepoImages = repoImages
	}

	var repoImagesToCopy []*image.Info
FindingImagesToCopy:
	for _, sourceImgInfo := range existingSourceRepoImages {
		for _, destImgInfo := range existingDestinationRepoImages {
			if sourceImgInfo.Signature == destImgInfo.Signature && sourceImgInfo.UniqueID == destImgInfo.UniqueID {
				continue FindingImagesToCopy
			}
		}
		repoImagesToCopy = append(repoImagesToCopy, sourceImgInfo)
	}

	logboek.Default.LogFDetails("Stages to copy: %d\n", len(repoImagesToCopy))

	maxWorkers := 10
	resultsChan := make(chan struct {
		error
		*image.Info
	}, 1000)
	jobsChan := make(chan *image.Info, 1000)

	for w := 0; w < maxWorkers; w++ {
		go runRopyWorker(projectName, fromStagesStorage, toStagesStorage, containerRuntime, w, jobsChan, resultsChan)
	}

	for _, imgInfo := range repoImagesToCopy {
		jobsChan <- imgInfo
	}
	close(jobsChan)

	failedCounter := 0
	succeededCounter := 0
	for i := 0; i < len(repoImagesToCopy); i++ {
		desc := <-resultsChan

		if desc.error != nil {
			failedCounter++
			logboek.LogErrorF("%5d/%d failed\n", failedCounter, len(repoImagesToCopy))
			errors = append(errors, desc.error)
		} else {
			succeededCounter++
			logboek.Default.LogF("%5d/%d synced\n", succeededCounter, len(repoImagesToCopy))
		}
	}

	if len(errors) > 0 {
		logboek.Default.LogLn()
		logboek.Default.LogFHighlight("synced %d/%d, failed %d/%d\n", succeededCounter, len(repoImagesToCopy), failedCounter, len(repoImagesToCopy))

		errorMsg := fmt.Sprintf("following errors occured:\n")
		for _, err := range errors {
			errorMsg += fmt.Sprintf(" - %s\n", err)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	isOk = true
	return nil
}

func runRopyWorker(projectName string, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime, workerId int, jobs chan *image.Info, results chan struct {
	error
	*image.Info
}) {
	for imgInfo := range jobs {
		results <- struct {
			error
			*image.Info
		}{
			copyStage(projectName, imgInfo, fromStagesStorage, toStagesStorage, containerRuntime),
			imgInfo,
		}
	}
}

func copyStage(projectName string, repoImage *image.Info, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime) error {
	img := container_runtime.NewStageImage(nil, repoImage.Name, containerRuntime.(*container_runtime.LocalDockerServerRuntime))

	logboek.Info.LogF("Fetching %s\n", img.Name())
	if err := fromStagesStorage.FetchImage(&container_runtime.DockerImage{Image: img}); err != nil {
		return fmt.Errorf("unable to fetch %s from %s: %s", repoImage.Name, fromStagesStorage.String(), err)
	}

	newImageName := toStagesStorage.ConstructStageImageName(projectName, repoImage.Signature, repoImage.UniqueID)
	logboek.Info.LogF("Renaming image %s to %s\n", img.Name(), newImageName)
	if err := containerRuntime.RenameImage(&container_runtime.DockerImage{Image: img}, newImageName); err != nil {
		return err
	}

	logboek.Info.LogF("Storing %s\n", newImageName)
	if err := toStagesStorage.StoreImage(&container_runtime.DockerImage{Image: img}); err != nil {
		return fmt.Errorf("unable to store %s to %s: %s", repoImage.Name, toStagesStorage.String(), err)
	}

	//deleteOpts := storage.DeleteRepoImageOptions{RmiForce: true, RmForce: true, RmContainersThatUseImage: true}
	//logboek.Default.LogF("Removing %s\n", repoImage.Name)
	//if err := fromStagesStorage.DeleteRepoImage(deleteOpts, repoImage); err != nil {
	//	return fmt.Errorf("unable to remove %s from %s: %s", repoImage.Name, fromStagesStorage.String(), err)
	//}

	return nil
}
