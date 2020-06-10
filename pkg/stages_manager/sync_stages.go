package stages_manager

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
)

type SyncStagesOptions struct {
	RemoveSource      bool
	CleanupLocalCache bool
	WithoutLock       bool
}

// SyncStages will make sure, that destination stages storage contains all stages from source stages storage.
// Repeatedly calling SyncStages will copy stages from source stages storage to destination, that already exists in the destination.
// SyncStages will not delete excess stages from destination storage, that does not exists in the source.
func SyncStages(projectName string, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, storageLockManager storage.LockManager, containerRuntime container_runtime.ContainerRuntime, opts SyncStagesOptions) error {
	isOk := false
	logboek.Default.LogProcessStart(fmt.Sprintf("Sync %q project stages", projectName), logboek.LevelLogProcessStartOptions{})
	defer func() {
		if isOk {
			logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
		} else {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		}
	}()

	if !opts.WithoutLock {
		if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: true}); err != nil {
			return fmt.Errorf("unable to lock stages and images of project %q: %s", projectName, err)
		} else {
			defer storageLockManager.Unlock(lock)
		}
	}

	logboek.Default.LogFDetails("Source      — %s\n", fromStagesStorage.String())
	logboek.Default.LogFDetails("Destination — %s\n", toStagesStorage.String())
	logboek.Default.LogOptionalLn()

	var errors []error

	getAllStagesFunc := func(logProcessMsg string, stagesStorage storage.StagesStorage) ([]image.StageID, error) {
		logboek.Default.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
		if stages, err := stagesStorage.GetAllStages(projectName); err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return nil, fmt.Errorf("unable to get repo images from %s: %s", fromStagesStorage.String(), err)
		} else {
			logboek.Default.LogFDetails("Stages count: %d\n", len(stages))
			logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})
			return stages, nil
		}
	}

	var existingSourceStages []image.StageID
	var existingDestinationStages []image.StageID

	if stages, err := getAllStagesFunc("Getting all repo images list from source stages storage", fromStagesStorage); err != nil {
		return fmt.Errorf("unable to get repo images from source %s: %s", fromStagesStorage.String(), err)
	} else {
		existingSourceStages = stages
	}

	if stages, err := getAllStagesFunc("Getting all repo images list from destination stages storage", toStagesStorage); err != nil {
		return fmt.Errorf("unable to get repo images from destination %s: %s", toStagesStorage.String(), err)
	} else {
		existingDestinationStages = stages
	}

	var stagesToSync []image.StageID

	for _, sourceStageDesc := range existingSourceStages {
		stageExistsInDestination := false
		for _, destStageDesc := range existingDestinationStages {
			if sourceStageDesc.Signature == destStageDesc.Signature && sourceStageDesc.UniqueID == destStageDesc.UniqueID {
				stageExistsInDestination = true
				break
			}
		}

		if !stageExistsInDestination || opts.RemoveSource {
			stagesToSync = append(stagesToSync, sourceStageDesc)
		}
	}

	logboek.Default.LogFDetails("Stages to sync: %d\n", len(stagesToSync))

	maxWorkers := 10
	resultsChan := make(chan struct {
		error
		image.StageID
	}, 1000)
	jobsChan := make(chan image.StageID, 1000)

	for w := 0; w < maxWorkers; w++ {
		go runSyncWorker(projectName, fromStagesStorage, toStagesStorage, containerRuntime, opts, w, jobsChan, resultsChan)
	}

	for _, stageDesc := range stagesToSync {
		jobsChan <- stageDesc
	}
	close(jobsChan)

	failedCounter := 0
	succeededCounter := 0
	for i := 0; i < len(stagesToSync); i++ {
		desc := <-resultsChan

		if desc.error != nil {
			failedCounter++
			logboek.LogErrorF("%5d/%d failed: %s\n", failedCounter, len(stagesToSync), desc.error)
			errors = append(errors, desc.error)
		} else {
			succeededCounter++
			logboek.Default.LogF("%5d/%d synced\n", succeededCounter, len(stagesToSync))
		}
	}

	if len(errors) > 0 {
		logboek.Default.LogLn()
		logboek.Default.LogFHighlight("synced %d/%d, failed %d/%d\n", succeededCounter, len(stagesToSync), failedCounter, len(stagesToSync))

		errorMsg := fmt.Sprintf("following errors occured:\n")
		for _, err := range errors {
			errorMsg += fmt.Sprintf(" - %s\n", err)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	isOk = true
	return nil
}

func runSyncWorker(projectName string, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime, opts SyncStagesOptions, workerId int, jobs chan image.StageID, results chan struct {
	error
	image.StageID
}) {
	for stageID := range jobs {
		results <- struct {
			error
			image.StageID
		}{
			syncStage(projectName, stageID, fromStagesStorage, toStagesStorage, containerRuntime, opts),
			stageID,
		}
	}
}

func syncStage(projectName string, stageID image.StageID, fromStagesStorage storage.StagesStorage, toStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime, opts SyncStagesOptions) error {
	if fromStagesStorage.Address() == storage.LocalStorageAddress || toStagesStorage.Address() == storage.LocalStorageAddress {
		opts.CleanupLocalCache = false
	}

	stageDesc, err := fromStagesStorage.GetStageDescription(projectName, stageID.Signature, stageID.UniqueID)
	if err != nil {
		return fmt.Errorf("error getting stage %s description from %s: %s", stageID.String(), fromStagesStorage.String(), err)
	} else if stageDesc == nil {
		// Bad stage id given: stage does not exists in the source stages storage
		return nil
	}

	if destStageDesc, err := toStagesStorage.GetStageDescription(projectName, stageID.Signature, stageID.UniqueID); err != nil {
		return fmt.Errorf("error getting stage %s description from %s: %s", stageID.String(), toStagesStorage.String(), err)
	} else if destStageDesc == nil {
		img := container_runtime.NewStageImage(nil, stageDesc.Info.Name, containerRuntime.(*container_runtime.LocalDockerServerRuntime))

		logboek.Info.LogF("Fetching %s\n", img.Name())
		if err := fromStagesStorage.FetchImage(&container_runtime.DockerImage{Image: img}); err != nil {
			return fmt.Errorf("unable to fetch %s from %s: %s", stageDesc.Info.Name, fromStagesStorage.String(), err)
		}

		newImageName := toStagesStorage.ConstructStageImageName(projectName, stageDesc.StageID.Signature, stageDesc.StageID.UniqueID)
		logboek.Info.LogF("Renaming image %s to %s\n", img.Name(), newImageName)
		if err := containerRuntime.RenameImage(&container_runtime.DockerImage{Image: img}, newImageName, opts.CleanupLocalCache); err != nil {
			return err
		}

		logboek.Info.LogF("Storing %s\n", newImageName)
		if err := toStagesStorage.StoreImage(&container_runtime.DockerImage{Image: img}); err != nil {
			return fmt.Errorf("unable to store %s to %s: %s", stageDesc.Info.Name, toStagesStorage.String(), err)
		}

		if opts.CleanupLocalCache {
			if err := containerRuntime.RemoveImage(&container_runtime.DockerImage{Image: img}); err != nil {
				return err
			}
		}
	}

	if opts.RemoveSource {
		deleteOpts := storage.DeleteImageOptions{RmiForce: true, RmForce: true, RmContainersThatUseImage: true}
		logboek.Info.LogF("Removing %s\n", stageDesc.Info.Name)
		if err := fromStagesStorage.DeleteStages(deleteOpts, stageDesc); err != nil {
			return fmt.Errorf("unable to remove %s from %s: %s", stageDesc.Info.Name, fromStagesStorage.String(), err)
		}
	}

	return nil
}
