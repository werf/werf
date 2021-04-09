package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/util/parallel"
	"github.com/werf/werf/pkg/werf"
)

var ErrShouldResetStagesStorageCache = errors.New("should reset storage cache")

func ShouldResetStagesStorageCache(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrShouldResetStagesStorageCache.Error())
	}
	return false
}

type StagesStorageManager struct {
	baseManager

	ProjectName string

	StorageLockManager storage.LockManager
	StagesStorage      storage.StagesStorage
	StagesStorageCache storage.StagesStorageCache

	SecondaryStagesStorageList []storage.StagesStorage

	// These will be released automatically when current process exits
	SharedHostImagesLocks []lockgate.LockHandle
}

func newStagesStorageManager(projectName string, stagesStorage storage.StagesStorage, secondaryStagesStorageList []storage.StagesStorage, storageLockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *StagesStorageManager {
	return &StagesStorageManager{
		ProjectName:        projectName,
		StorageLockManager: storageLockManager,
		StagesStorageCache: stagesStorageCache,

		StagesStorage:              stagesStorage,
		SecondaryStagesStorageList: secondaryStagesStorageList,
	}
}

func (m *StagesStorageManager) ResetStagesStorageCache(ctx context.Context) error {
	msg := fmt.Sprintf("Reset storage cache %s for project %q", m.StagesStorageCache.String(), m.ProjectName)
	return logboek.Context(ctx).Default().LogProcess(msg).DoError(func() error {
		return m.StagesStorageCache.DeleteAllStages(ctx, m.ProjectName)
	})
}

func (m *StagesStorageManager) GetStageDescriptionList(ctx context.Context) ([]*image.StageDescription, error) {
	stageIDs, err := m.StagesStorage.GetStagesIDs(ctx, m.ProjectName)
	if err != nil {
		return nil, err
	}

	var mutex sync.Mutex
	var stages []*image.StageDescription
	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		if stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, m.StagesStorage, getStageDescriptionOptions{StageShouldExist: false, WithManifestCache: m.getWithManifestCacheOption()}); err != nil {
			return err
		} else if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
		} else {
			mutex.Lock()
			defer mutex.Unlock()

			stages = append(stages, stageDesc)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return stages, nil
}

type ForEachDeleteStageOptions struct {
	storage.DeleteImageOptions
	storage.FilterStagesAndProcessRelatedDataOptions
}

func (m *StagesStorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error {
	var err error
	stagesDescriptions, err = m.StagesStorage.FilterStagesAndProcessRelatedData(ctx, stagesDescriptions, options.FilterStagesAndProcessRelatedDataOptions)
	if err != nil {
		return err
	}

	for _, stageDesc := range stagesDescriptions {
		if err := m.StagesStorageCache.DeleteStagesByDigest(ctx, m.ProjectName, stageDesc.StageID.Digest); err != nil {
			return fmt.Errorf("unable to delete storage cache record (%s): %s", stageDesc.StageID.Digest, err)
		}
	}

	return parallel.DoTasks(ctx, len(stagesDescriptions), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDescription := stagesDescriptions[taskId]
		err := m.StagesStorage.DeleteStage(ctx, stageDescription, options.DeleteImageOptions)
		return f(ctx, stageDescription, err)
	})
}

func (m *StagesStorageManager) LockStageImage(ctx context.Context, imageName string) error {
	imageLockName := container_runtime.ImageLockName(imageName)

	_, lock, err := werf.AcquireHostLock(ctx, imageLockName, lockgate.AcquireOptions{Shared: true})
	if err != nil {
		return fmt.Errorf("error locking %q shared lock: %s", imageLockName, err)
	}

	m.SharedHostImagesLocks = append(m.SharedHostImagesLocks, lock)

	return nil
}

func (m *StagesStorageManager) FetchStage(ctx context.Context, stg stage.Interface) error {
	if err := m.LockStageImage(ctx, stg.GetImage().Name()); err != nil {
		return fmt.Errorf("error locking stage image %q: %s", stg.GetImage().Name(), err)
	}

	logboek.Context(ctx).Debug().LogF("-- StagesManager.FetchStage %s\n", stg.LogDetailedName())
	if freshStageDescription, err := m.StagesStorage.GetStageDescription(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Digest, stg.GetImage().GetStageDescription().StageID.UniqueID); err == storage.ErrBrokenImage {
		logboek.Context(ctx).Error().LogF("Invalid stage %s image %q! Stage image is broken and is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), m.ProjectName)

		logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stg.GetImage().Name(), m.StagesStorage.String())
		if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Digest, stg.GetImage().GetStageDescription().StageID.UniqueID); err != nil {
			return fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
		}

		return ErrShouldResetStagesStorageCache
	} else if err != nil {
		return fmt.Errorf("unable to get stage %s description: %s", stg.GetImage().GetStageDescription().StageID.String(), err)
	} else if freshStageDescription == nil {
		logboek.Context(ctx).Error().LogF("Invalid stage %s image %q! Stage is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), m.ProjectName)
		return ErrShouldResetStagesStorageCache
	}

	if shouldFetch, err := m.StagesStorage.ShouldFetchImage(ctx, &container_runtime.DockerImage{Image: stg.GetImage()}); err == nil && shouldFetch {
		if err := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from storage", stg.LogDetailedName()).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				logboek.Context(ctx).Info().LogF("Image name: %s\n", stg.GetImage().Name())

				if err := m.StagesStorage.FetchImage(ctx, &container_runtime.DockerImage{Image: stg.GetImage()}); err == storage.ErrBrokenImage {
					logboek.Context(ctx).Error().LogF("Unable to fetch image %q: %s. Stages storage cache for project %q should be reset!\n", stg.GetImage().Name(), err, m.ProjectName)

					logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stg.GetImage().Name(), m.StagesStorage.String())
					if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Digest, stg.GetImage().GetStageDescription().StageID.UniqueID); err != nil {
						return fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
					}

					return ErrShouldResetStagesStorageCache
				} else if err != nil {
					return fmt.Errorf("unable to fetch stage %s image %s from storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
				}

				return nil
			}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, stg.GetImage().Name()); err != nil {
		return fmt.Errorf("error accessing last recently used images cache: %s", err)
	}

	return nil
}

func (m *StagesStorageManager) SelectSuitableStage(ctx context.Context, c stage.Conveyor, stg stage.Interface, stages []*image.StageDescription) (*image.StageDescription, error) {
	if len(stages) == 0 {
		return nil, nil
	}

	var stageDesc *image.StageDescription
	if err := logboek.Context(ctx).Info().LogProcess("Selecting suitable image for stage %s by digest %s", stg.Name(), stg.GetDigest()).
		DoError(func() error {
			var err error
			stageDesc, err = stg.SelectSuitableStage(ctx, c, stages)
			return err
		}); err != nil {
		return nil, err
	}
	if stageDesc == nil {
		return nil, nil
	}

	imgInfoData, err := yaml.Marshal(stageDesc)
	if err != nil {
		panic(err)
	}

	logboek.Context(ctx).Debug().LogBlock("Selected cache image").
		Options(func(options types.LogBlockOptionsInterface) {
			options.Style(style.Highlight())
		}).
		Do(func() {
			logboek.Context(ctx).Debug().LogF(string(imgInfoData))
		})

	return stageDesc, nil
}

func (m *StagesStorageManager) AtomicStoreStagesByDigestToCache(ctx context.Context, stageName, stageDigest string, stageIDs []image.StageID) error {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageDigest); err != nil {
		return fmt.Errorf("error locking stage %s cache by digest %s: %s", stageName, stageDigest, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	return logboek.Context(ctx).Info().LogProcess("Storing stage %s images by digest %s into storage cache", stageName, stageDigest).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesByDigest(ctx, m.ProjectName, stageDigest, stageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by digest %s into storage cache: %s", stageName, stageDigest, err)
			}
			return nil
		})
}

func (m *StagesStorageManager) GetStagesByDigest(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error) {
	cacheExists, cacheStages, err := m.getStagesByDigestFromCache(ctx, stageName, stageDigest)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s stages by digest %q from cache: %s", stageName, stageDigest, err)
	}
	if cacheExists {
		return cacheStages, nil
	}

	logboek.Context(ctx).Default().LogF(
		"Stage %s cache by digest %s is not exists in the stages storage cache: will request fresh stages from storage and set stages storage cache by digest %s\n",
		stageName, stageDigest, stageDigest,
	)
	return m.atomicGetStagesByDigestWithStagesStorageCacheStore(ctx, stageName, stageDigest)
}

func (m *StagesStorageManager) GetStagesByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error) {
	stageIDs, err := m.getStagesIDsByDigestFromStagesStorage(ctx, stageName, stageDigest, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("unable to get stages ids from %s by digest %s for stage %s: %s", stagesStorage.String(), stageDigest, stageName, err)
	}

	stages, err := m.getStagesDescriptions(ctx, stageIDs, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage descriptions by ids from %s: %s", stagesStorage.String(), err)
	}

	return stages, nil
}

func (m *StagesStorageManager) CopySuitableByDigestStage(ctx context.Context, stageDesc *image.StageDescription, sourceStagesStorage, destinationStagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime) (*image.StageDescription, error) {
	img := container_runtime.NewStageImage(nil, stageDesc.Info.Name, containerRuntime.(*container_runtime.LocalDockerServerRuntime))

	logboek.Context(ctx).Info().LogF("Fetching %s\n", img.Name())
	if err := sourceStagesStorage.FetchImage(ctx, &container_runtime.DockerImage{Image: img}); err != nil {
		return nil, fmt.Errorf("unable to fetch %s from %s: %s", stageDesc.Info.Name, sourceStagesStorage.String(), err)
	}

	newImageName := destinationStagesStorage.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.UniqueID)
	logboek.Context(ctx).Info().LogF("Renaming image %s to %s\n", img.Name(), newImageName)
	if err := containerRuntime.RenameImage(ctx, &container_runtime.DockerImage{Image: img}, newImageName, false); err != nil {
		return nil, err
	}

	logboek.Context(ctx).Info().LogF("Storing %s\n", newImageName)
	if err := destinationStagesStorage.StoreImage(ctx, &container_runtime.DockerImage{Image: img}); err != nil {
		return nil, fmt.Errorf("unable to store %s to %s: %s", stageDesc.Info.Name, destinationStagesStorage.String(), err)
	}

	if destinationStageDesc, err := getStageDescription(ctx, m.ProjectName, *stageDesc.StageID, destinationStagesStorage, getStageDescriptionOptions{StageShouldExist: true, WithManifestCache: m.getWithManifestCacheOption()}); err != nil {
		return nil, fmt.Errorf("unable to get stage %s description from %s: %s", stageDesc.StageID.String(), destinationStagesStorage.String(), err)
	} else {
		return destinationStageDesc, nil
	}
}

func (m *StagesStorageManager) getWithManifestCacheOption() bool {
	return m.StagesStorage.Address() != storage.LocalStorageAddress
}

func (m *StagesStorageManager) getStagesByDigestFromCache(ctx context.Context, stageName, stageDigest string) (bool, []*image.StageDescription, error) {
	var cacheExists bool
	var cacheStagesIDs []image.StageID

	err := logboek.Context(ctx).Info().LogProcess("Getting stage %s images by digest %s from storage cache", stageName, stageDigest).
		DoError(func() error {
			var err error
			cacheExists, cacheStagesIDs, err = m.StagesStorageCache.GetStagesByDigest(ctx, m.ProjectName, stageDigest)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from storage cache: %s", m.ProjectName, stageDigest, err)
			}
			return nil
		})

	var stages []*image.StageDescription

	for _, stageID := range cacheStagesIDs {
		if stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, m.StagesStorage, getStageDescriptionOptions{StageShouldExist: true, WithManifestCache: m.getWithManifestCacheOption()}); err != nil {
			return false, nil, fmt.Errorf("unable to get stage %q description: %s", stageID.String(), err)
		} else {
			stages = append(stages, stageDesc)
		}
	}

	return cacheExists, stages, err
}

func (m *StagesStorageManager) getStagesIDsByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]image.StageID, error) {
	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Info().LogProcess("Get %s stages by digest %s from storage", stageName, stageDigest).
		DoError(func() error {
			var err error
			stageIDs, err = stagesStorage.GetStagesIDsByDigest(ctx, m.ProjectName, stageDigest)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from storage: %s", m.StagesStorage.String(), stageDigest, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	return stageIDs, nil
}

func (m *StagesStorageManager) getStagesDescriptions(ctx context.Context, stageIDs []image.StageID, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error) {
	var stages []*image.StageDescription
	for _, stageID := range stageIDs {
		if stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, stagesStorage, getStageDescriptionOptions{StageShouldExist: false, WithManifestCache: m.getWithManifestCacheOption()}); err != nil {
			return nil, err
		} else if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
			continue
		} else {
			stages = append(stages, stageDesc)
		}
	}

	return stages, nil
}

func (m *StagesStorageManager) atomicGetStagesByDigestWithStagesStorageCacheStore(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error) {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageDigest); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", m.ProjectName, stageDigest, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	stageIDs, err := m.getStagesIDsByDigestFromStagesStorage(ctx, stageName, stageDigest, m.StagesStorage)
	if err != nil {
		return nil, fmt.Errorf("unable to get stages ids from %s by digest %s for stage %s: %s", m.StagesStorage.String(), stageDigest, stageName, err)
	}

	validStages, err := m.getStagesDescriptions(ctx, stageIDs, m.StagesStorage)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage descriptions by ids from %s: %s", m.StagesStorage.String(), err)
	}

	var validStageIDs []image.StageID
	for _, stage := range validStages {
		validStageIDs = append(validStageIDs, *stage.StageID)
	}

	if err := logboek.Context(ctx).Info().LogProcess("Storing %s stages by digest %s into stages storage cache", stageName, stageDigest).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesByDigest(ctx, m.ProjectName, stageDigest, validStageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by digest %s into storage cache: %s", stageName, stageDigest, err)
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return validStages, nil
}

type getStageDescriptionOptions struct {
	StageShouldExist  bool
	WithManifestCache bool
}

func getStageDescription(ctx context.Context, projectName string, stageID image.StageID, stagesStorage storage.StagesStorage, opts getStageDescriptionOptions) (*image.StageDescription, error) {
	stageImageName := stagesStorage.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)

	if opts.WithManifestCache {
		logboek.Context(ctx).Debug().LogF("Getting image %s info from manifest cache...\n", stageImageName)
		if imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, stagesStorage.String(), stageImageName); err != nil {
			return nil, fmt.Errorf("error getting image %s info from manifest cache: %s", stageImageName, err)
		} else if imgInfo != nil {
			logboek.Context(ctx).Debug().LogF("Got image %s info from manifest cache (CACHE HIT)\n", stageImageName)
			return &image.StageDescription{
				StageID: &image.StageID{Digest: stageID.Digest, UniqueID: stageID.UniqueID},
				Info:    imgInfo,
			}, nil
		} else {
			logboek.Context(ctx).Info().LogF("Not found %s image info in manifest cache (CACHE MISS)\n", stageImageName)
		}
	}

	logboek.Context(ctx).Debug().LogF("Getting digest %q uniqueID %d stage info from %s...\n", stageID.Digest, stageID.UniqueID, stagesStorage.String())
	if stageDesc, err := stagesStorage.GetStageDescription(ctx, projectName, stageID.Digest, stageID.UniqueID); err == storage.ErrBrokenImage {
		if opts.StageShouldExist {
			logboek.Context(ctx).Error().LogF("Invalid stage image %q! Stage is broken and is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stageImageName, stagesStorage.String(), projectName)

			logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stageImageName, stagesStorage.String())
			if err := stagesStorage.RejectStage(ctx, projectName, stageID.Digest, stageID.UniqueID); err != nil {
				return nil, fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %s", stageID.String(), stageImageName, stagesStorage.String(), err)
			}

			return nil, ErrShouldResetStagesStorageCache
		}

		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error getting digest %q uniqueID %d stage info from %s: %s", stageID.Digest, stageID.UniqueID, stagesStorage.String(), err)
	} else if stageDesc != nil {
		if opts.WithManifestCache {
			logboek.Context(ctx).Debug().LogF("Storing image %s info into manifest cache\n", stageImageName)
			if err := image.CommonManifestCache.StoreImageInfo(ctx, stagesStorage.String(), stageDesc.Info); err != nil {
				return nil, fmt.Errorf("unable to store image %s info into manifest cache: %s", stageImageName, err)
			}
		}

		return stageDesc, nil
	} else if opts.StageShouldExist {
		logboek.Context(ctx).Error().LogF("Invalid stage image %q! Stage is no longer available in the %s. Storage cache for project %q should be reset!\n", stageImageName, stagesStorage.String(), projectName)
		return nil, ErrShouldResetStagesStorageCache
	} else {
		return nil, nil
	}
}

func (m *StagesStorageManager) GenerateStageUniqueID(digest string, stages []*image.StageDescription) (string, int64) {
	var imageName string

	for {
		timeNow := time.Now().UTC()
		uniqueID := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)
		imageName = m.StagesStorage.ConstructStageImageName(m.ProjectName, digest, uniqueID)

		for _, stageDesc := range stages {
			if stageDesc.Info.Name == imageName {
				continue
			}
		}
		return imageName, uniqueID
	}
}

type rmImageMetadataTask struct {
	commit  string
	stageID string
}

func (m *StagesStorageManager) ForEachRmImageMetadata(ctx context.Context, projectName, imageNameOrID string, stageIDCommitList map[string][]string, f func(ctx context.Context, commit, stageID string, err error) error) error {
	var tasks []rmImageMetadataTask
	for stageID, commitList := range stageIDCommitList {
		for _, commit := range commitList {
			tasks = append(tasks, rmImageMetadataTask{
				commit:  commit,
				stageID: stageID,
			})
		}
	}

	return parallel.DoTasks(ctx, len(tasks), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		task := tasks[taskId]
		err := m.StagesStorage.RmImageMetadata(ctx, projectName, imageNameOrID, task.commit, task.stageID)
		return f(ctx, task.commit, task.stageID, err)
	})
}

func (m *StagesStorageManager) ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error {
	return parallel.DoTasks(ctx, len(managedImages), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		managedImage := managedImages[taskId]
		err := m.StagesStorage.RmManagedImage(ctx, projectName, managedImage)
		return f(ctx, managedImage, err)
	})
}

func (m *StagesStorageManager) ForEachGetImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.ImportMetadata, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]
		metadata, err := m.StagesStorage.GetImportMetadata(ctx, projectName, id)
		return f(ctx, id, metadata, err)
	})
}

func (m *StagesStorageManager) ForEachRmImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, id string, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]
		err := m.StagesStorage.RmImportMetadata(ctx, projectName, id)
		return f(ctx, id, err)
	})
}
