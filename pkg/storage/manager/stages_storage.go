package manager

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

var ErrShouldResetStagesStorageCache = errors.New("should reset stages storage cache")

func ShouldResetStagesStorageCache(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrShouldResetStagesStorageCache.Error())
	}
	return false
}

type StagesStorageManager struct {
	baseManager

	StagesSwitchFromLocalBlockDir string
	ProjectName                   string

	StorageLockManager storage.LockManager
	StagesStorage      storage.StagesStorage
	StagesStorageCache storage.StagesStorageCache

	// These will be released automatically when current process exits
	SharedHostImagesLocks []lockgate.LockHandle
}

func newStagesStorageManager(projectName string, storageLockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *StagesStorageManager {
	return &StagesStorageManager{
		StagesSwitchFromLocalBlockDir: filepath.Join(werf.GetServiceDir(), "stages_switch_from_local_block"),
		ProjectName:                   projectName,
		StorageLockManager:            storageLockManager,
		StagesStorageCache:            stagesStorageCache,
	}
}

func (m *StagesStorageManager) ResetStagesStorageCache(ctx context.Context) error {
	msg := fmt.Sprintf("Reset stages storage cache %s for project %q", m.StagesStorageCache.String(), m.ProjectName)
	return logboek.Context(ctx).Default().LogProcess(msg).DoError(func() error {
		return m.StagesStorageCache.DeleteAllStages(ctx, m.ProjectName)
	})
}

func (m *StagesStorageManager) getStagesSwitchFromLocalBlock() (string, error) {
	f := filepath.Join(m.StagesSwitchFromLocalBlockDir, m.ProjectName)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("error accessing %s: %s", f, err)
	}

	if dataBytes, err := ioutil.ReadFile(f); err != nil {
		return "", fmt.Errorf("error reading %s: %s", f, err)
	} else {
		return strings.TrimSpace(string(dataBytes)), nil
	}
}

func (m *StagesStorageManager) checkStagesSwitchFromLocalBlock(stagesStorageAddress string) error {
	if switchFromLocalBlock, err := m.getStagesSwitchFromLocalBlock(); err != nil {
		return err
	} else if switchFromLocalBlock != "" && stagesStorageAddress == storage.LocalStorageAddress {
		return fmt.Errorf(
			`Project %q stages storage has been switched from %s to %s!

 1. Remove --stages-storage=%s param if it is specified explicitly.
 2. If 'werf ci-env' command is used, then WERF_STAGES_STORAGE already should be exported — make sure that WERF_STAGES_STORAGE equals %s in this case.
 3. Otherwise explicitly specify --stages-storage=%s (or export WERF_STAGES_STORAGE=%s).`,
			m.ProjectName,
			storage.LocalStorageAddress,
			switchFromLocalBlock,
			storage.LocalStorageAddress,
			switchFromLocalBlock,
			switchFromLocalBlock,
			switchFromLocalBlock,
		)
	}

	return nil
}

func (m *StagesStorageManager) writeStagesSwitchFromLocalBlock(stagesStorageAddress string) error {
	f := filepath.Join(m.StagesSwitchFromLocalBlockDir, m.ProjectName)
	d := filepath.Dir(f)
	if err := os.MkdirAll(d, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %s: %s", d, err)
	}
	if err := ioutil.WriteFile(f, []byte(fmt.Sprintf("%s\n", stagesStorageAddress)), 0644); err != nil {
		return fmt.Errorf("error writing %s: %s", f, err)
	}
	return nil
}

func stagesSwitchFromLocalBlockLockName(projectName string) string {
	return fmt.Sprintf("stages_switch_from_local_block.%s", projectName)
}

func (m *StagesStorageManager) SetStagesSwitchFromLocalBlock(ctx context.Context, newStagesStorage storage.StagesStorage) error {
	if _, lock, err := werf.AcquireHostLock(ctx, stagesSwitchFromLocalBlockLockName(m.ProjectName), lockgate.AcquireOptions{}); err != nil {
		return err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if err := m.writeStagesSwitchFromLocalBlock(newStagesStorage.Address()); err != nil {
		return err
	}
	return nil
}

func (m *StagesStorageManager) UseStagesStorage(ctx context.Context, stagesStorage storage.StagesStorage) error {
	if _, lock, err := werf.AcquireHostLock(ctx, stagesSwitchFromLocalBlockLockName(m.ProjectName), lockgate.AcquireOptions{}); err != nil {
		return err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if err := m.checkStagesSwitchFromLocalBlock(stagesStorage.Address()); err != nil {
		return err
	}
	m.StagesStorage = stagesStorage
	return nil
}

func (m *StagesStorageManager) GetAllStages(ctx context.Context) ([]*image.StageDescription, error) {
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

		if stageDesc, err := m.getStageDescription(ctx, stageID, getStageDescriptionOptions{StageShouldExist: false, WithManifestCache: true}); err != nil {
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

func (m *StagesStorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(stageDesc *image.StageDescription, err error) error) error {
	var err error
	stagesDescriptions, err = m.StagesStorage.FilterStagesAndProcessRelatedData(ctx, stagesDescriptions, options.FilterStagesAndProcessRelatedDataOptions)
	if err != nil {
		return err
	}

	for _, stageDesc := range stagesDescriptions {
		if err := m.StagesStorageCache.DeleteStagesBySignature(ctx, m.ProjectName, stageDesc.StageID.Signature); err != nil {
			return fmt.Errorf("unable to delete stages storage cache record (%s): %s", stageDesc.StageID.Signature, err)
		}
	}

	return parallel.DoTasks(ctx, len(stagesDescriptions), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageDescription := stagesDescriptions[taskId]
		err := m.StagesStorage.DeleteStage(ctx, stageDescription, options.DeleteImageOptions)
		return f(stageDescription, err)
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
	if freshStageDescription, err := m.StagesStorage.GetStageDescription(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Signature, stg.GetImage().GetStageDescription().StageID.UniqueID); err == storage.ErrBrokenImage {
		logboek.Context(ctx).Error().LogF("Invalid stage %s image %q! Stage image is broken and is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), m.ProjectName)

		logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stg.GetImage().Name(), m.StagesStorage.String())
		if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Signature, stg.GetImage().GetStageDescription().StageID.UniqueID); err != nil {
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
		if err := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from stages storage", stg.LogDetailedName()).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				logboek.Context(ctx).Info().LogF("Image name: %s\n", stg.GetImage().Name())
				if err := m.StagesStorage.FetchImage(ctx, &container_runtime.DockerImage{Image: stg.GetImage()}); err == storage.ErrBrokenImage {
					logboek.Context(ctx).Error().LogF("Unable to fetch image %q: %s. Stages storage cache for project %q should be reset!\n", stg.GetImage().Name(), err, m.ProjectName)

					logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stg.GetImage().Name(), m.StagesStorage.String())
					if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Signature, stg.GetImage().GetStageDescription().StageID.UniqueID); err != nil {
						return fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
					}

					return ErrShouldResetStagesStorageCache
				} else if err != nil {
					return fmt.Errorf("unable to fetch stage %s image %s from stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
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
	if err := logboek.Context(ctx).Info().LogProcess("Selecting suitable image for stage %s by signature %s", stg.Name(), stg.GetSignature()).
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

func (m *StagesStorageManager) AtomicStoreStagesBySignatureToCache(ctx context.Context, stageName, stageSig string, stageIDs []image.StageID) error {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageSig); err != nil {
		return fmt.Errorf("error locking stage %s cache by signature %s: %s", stageName, stageSig, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	return logboek.Context(ctx).Info().LogProcess("Storing stage %s images by signature %s into stages storage cache", stageName, stageSig).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesBySignature(ctx, m.ProjectName, stageSig, stageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		})
}

func (m *StagesStorageManager) GetStagesBySignature(ctx context.Context, stageName, stageSig string) ([]*image.StageDescription, error) {
	cacheExists, cacheStages, err := m.getStagesBySignatureFromCache(ctx, stageName, stageSig)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s stages by signature %q from cache: %s", stageName, stageSig, err)
	}
	if cacheExists {
		return cacheStages, nil
	}

	logboek.Context(ctx).Default().LogF(
		"Stage %s cache by signature %s is not exists in the stages storage cache: will request fresh stages from stages storage and set stages storage cache by signature %s\n",
		stageName, stageSig, stageSig,
	)
	return m.atomicGetStagesBySignatureWithStagesStorageCacheStore(ctx, stageName, stageSig)
}

func (m *StagesStorageManager) getStagesBySignatureFromCache(ctx context.Context, stageName, stageSig string) (bool, []*image.StageDescription, error) {
	var cacheExists bool
	var cacheStagesIDs []image.StageID

	err := logboek.Context(ctx).Info().LogProcess("Getting stage %s images by signature %s from stages storage cache", stageName, stageSig).
		DoError(func() error {
			var err error
			cacheExists, cacheStagesIDs, err = m.StagesStorageCache.GetStagesBySignature(ctx, m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage cache: %s", m.ProjectName, stageSig, err)
			}
			return nil
		})

	var stages []*image.StageDescription

	for _, stageID := range cacheStagesIDs {
		if stageDesc, err := m.getStageDescription(ctx, stageID, getStageDescriptionOptions{StageShouldExist: true, WithManifestCache: true}); err != nil {
			return false, nil, fmt.Errorf("unable to get stage %q description: %s", stageID.String(), err)
		} else {
			stages = append(stages, stageDesc)
		}
	}

	return cacheExists, stages, err
}

func (m *StagesStorageManager) atomicGetStagesBySignatureWithStagesStorageCacheStore(ctx context.Context, stageName, stageSig string) ([]*image.StageDescription, error) {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", m.ProjectName, stageSig, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Default().LogProcess("Get %s stages by signature %s from stages storage", stageName, stageSig).
		DoError(func() error {
			var err error
			stageIDs, err = m.StagesStorage.GetStagesIDsBySignature(ctx, m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", m.StagesStorage.String(), stageSig, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	var validStages []*image.StageDescription
	var validStageIDs []image.StageID
	for _, stageID := range stageIDs {
		if stageDesc, err := m.getStageDescription(ctx, stageID, getStageDescriptionOptions{StageShouldExist: false, WithManifestCache: true}); err != nil {
			return nil, err
		} else if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
			continue
		} else {
			validStages = append(validStages, stageDesc)
			validStageIDs = append(validStageIDs, stageID)
		}
	}

	if err := logboek.Context(ctx).Info().LogProcess("Storing %s stages by signature %s into stages storage cache", stageName, stageSig).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesBySignature(ctx, m.ProjectName, stageSig, validStageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by signature %s into stages storage cache: %s", stageName, stageSig, err)
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

func (m *StagesStorageManager) getStageDescription(ctx context.Context, stageID image.StageID, opts getStageDescriptionOptions) (*image.StageDescription, error) {
	stageImageName := m.StagesStorage.ConstructStageImageName(m.ProjectName, stageID.Signature, stageID.UniqueID)

	if opts.WithManifestCache {
		logboek.Context(ctx).Debug().LogF("Getting image %s info from manifest cache...\n", stageImageName)
		if imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, m.StagesStorage.String(), stageImageName); err != nil {
			return nil, fmt.Errorf("error getting image %s info from manifest cache: %s", stageImageName, err)
		} else if imgInfo != nil {
			logboek.Context(ctx).Debug().LogF("Got image %s info from manifest cache (CACHE HIT)\n", stageImageName)
			return &image.StageDescription{
				StageID: &image.StageID{Signature: stageID.Signature, UniqueID: stageID.UniqueID},
				Info:    imgInfo,
			}, nil
		} else {
			logboek.Context(ctx).Info().LogF("Not found %s image info in manifest cache (CACHE MISS)\n", stageImageName)
		}
	}

	logboek.Context(ctx).Debug().LogF("Getting signature %q uniqueID %d stage info from %s...\n", stageID.Signature, stageID.UniqueID, m.StagesStorage.String())
	if stageDesc, err := m.StagesStorage.GetStageDescription(ctx, m.ProjectName, stageID.Signature, stageID.UniqueID); err == storage.ErrBrokenImage {
		logboek.Context(ctx).Error().LogF("Invalid stage image %q! Stage is broken and is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stageImageName, m.StagesStorage.String(), m.ProjectName)

		logboek.Context(ctx).Error().LogF("Will mark image %q as rejected in the stages storage %s\n", stageImageName, m.StagesStorage.String())
		if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stageID.Signature, stageID.UniqueID); err != nil {
			return nil, fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %s", stageID.String(), stageImageName, m.StagesStorage.String(), err)
		}

		return nil, ErrShouldResetStagesStorageCache
	} else if err != nil {
		return nil, fmt.Errorf("error getting signature %q uniqueID %d stage info from %s: %s", stageID.Signature, stageID.UniqueID, m.StagesStorage.String(), err)
	} else if stageDesc != nil {
		logboek.Context(ctx).Debug().LogF("Storing image %s info into manifest cache\n", stageImageName)
		if err := image.CommonManifestCache.StoreImageInfo(ctx, m.StagesStorage.String(), stageDesc.Info); err != nil {
			return nil, fmt.Errorf("error storing image %s info into manifest cache: %s", stageImageName, err)
		}

		return stageDesc, nil
	} else if opts.StageShouldExist {
		logboek.Context(ctx).Error().LogF("Invalid stage image %q! Stage is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stageImageName, m.StagesStorage.String(), m.ProjectName)
		return nil, ErrShouldResetStagesStorageCache
	} else {
		return nil, nil
	}
}

func (m *StagesStorageManager) GenerateStageUniqueID(signature string, stages []*image.StageDescription) (string, int64) {
	var imageName string

	for {
		timeNow := time.Now().UTC()
		uniqueID := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)
		imageName = m.StagesStorage.ConstructStageImageName(m.ProjectName, signature, uniqueID)

		for _, stageDesc := range stages {
			if stageDesc.Info.Name == imageName {
				continue
			}
		}
		return imageName, uniqueID
	}
}

func (m *StagesStorageManager) ForEachGetImageMetadataByCommit(ctx context.Context, projectName, imageName string, f func(commit string, imageMetadata *storage.ImageMetadata, err error) error) error {
	commits, err := m.StagesStorage.GetImageCommits(ctx, projectName, imageName)
	if err != nil {
		return fmt.Errorf("get image %s commits failed: %s", imageName, err)
	}

	return parallel.DoTasks(ctx, len(commits), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		commit := commits[taskId]

		imageMetadata, err := storage.GetImageMetadataCache().GetImageMetadata(ctx, m.StagesStorage.Address(), imageName, commit)
		if err != nil {
			return fmt.Errorf("get image metadata failed: %s", err)
		}

		if imageMetadata == nil {
			logboek.Context(ctx).Info().LogF("Not found image metadata image %s commit %s in cache (CACHE MISS)\n", imageName, commit)

			imageMetadata, err = m.StagesStorage.GetImageMetadataByCommit(ctx, projectName, imageName, commit)
			if err != nil {
				return f(commit, imageMetadata, err)
			}

			err = storage.GetImageMetadataCache().StoreImageMetadata(ctx, m.StagesStorage.Address(), imageName, commit, imageMetadata)
			if err != nil {
				return fmt.Errorf("store image metadata failed: %s", err)
			}
		} else {
			logboek.Context(ctx).Info().LogF("Got image metadata image %s commit from cache (CACHE HIT)\n", imageName, commit)
		}

		return f(commit, imageMetadata, nil)
	})
}

func (m *StagesStorageManager) ForEachRmImageCommit(ctx context.Context, projectName, imageName string, commits []string, f func(commit string, err error) error) error {
	return parallel.DoTasks(ctx, len(commits), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		commit := commits[taskId]
		err := m.StagesStorage.RmImageCommit(ctx, projectName, imageName, commit)
		return f(commit, err)
	})
}
