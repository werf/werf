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

 1. Remove --repo=%s param if it is specified explicitly.
 2. If 'werf ci-env' command is used, then WERF_REPO already should be exported — make sure that WERF_REPO equals %s in this case.
 3. Otherwise explicitly specify --repo=%s (or export WERF_REPO=%s).`,
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

func (m *StagesStorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error {
	var err error
	stagesDescriptions, err = m.StagesStorage.FilterStagesAndProcessRelatedData(ctx, stagesDescriptions, options.FilterStagesAndProcessRelatedDataOptions)
	if err != nil {
		return err
	}

	for _, stageDesc := range stagesDescriptions {
		if err := m.StagesStorageCache.DeleteStagesByDigest(ctx, m.ProjectName, stageDesc.StageID.Digest); err != nil {
			return fmt.Errorf("unable to delete stages storage cache record (%s): %s", stageDesc.StageID.Digest, err)
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

func (m *StagesStorageManager) FetchStage(ctx context.Context, stg stage.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- StagesManager.FetchStage %s\n", stg.LogDetailedName())
	if freshStageDescription, err := m.StagesStorage.GetStageDescription(ctx, m.ProjectName, stg.GetImage().GetStageDescription().StageID.Digest, stg.GetImage().GetStageDescription().StageID.UniqueID); err != nil {
		return err
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
				if err := m.StagesStorage.FetchImage(ctx, &container_runtime.DockerImage{Image: stg.GetImage()}); err != nil {
					return fmt.Errorf("unable to fetch stage %s image %s from stages storage %s: %s", stg.LogDetailedName(), stg.GetImage().Name(), m.StagesStorage.String(), err)
				}
				return nil
			}); err != nil {
			return err
		}
	} else if err != nil {
		return err
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

func (m *StagesStorageManager) AtomicStoreStagesByDigestToCache(ctx context.Context, stageName, stageSig string, stageIDs []image.StageID) error {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageSig); err != nil {
		return fmt.Errorf("error locking stage %s cache by digest %s: %s", stageName, stageSig, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	return logboek.Context(ctx).Info().LogProcess("Storing stage %s images by digest %s into stages storage cache", stageName, stageSig).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesByDigest(ctx, m.ProjectName, stageSig, stageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by digest %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		})
}

func (m *StagesStorageManager) GetStagesByDigest(ctx context.Context, stageName, stageSig string) ([]*image.StageDescription, error) {
	cacheExists, cacheStages, err := m.getStagesByDigestFromCache(ctx, stageName, stageSig)
	if err != nil {
		return nil, err
	}
	if cacheExists {
		return cacheStages, nil
	}

	logboek.Context(ctx).Info().LogF(
		"Stage %s cache by digest %s is not exists in the stages storage cache: resetting stages storage cache\n",
		stageName, stageSig,
	)
	return m.atomicGetStagesByDigestWithCacheReset(ctx, stageName, stageSig)
}

func (m *StagesStorageManager) getStagesByDigestFromCache(ctx context.Context, stageName, stageSig string) (bool, []*image.StageDescription, error) {
	var cacheExists bool
	var cacheStagesIDs []image.StageID

	err := logboek.Context(ctx).Info().LogProcess("Getting stage %s images by digest %s from stages storage cache", stageName, stageSig).
		DoError(func() error {
			var err error
			cacheExists, cacheStagesIDs, err = m.StagesStorageCache.GetStagesByDigest(ctx, m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage cache: %s", m.ProjectName, stageSig, err)
			}
			return nil
		})

	var stages []*image.StageDescription

	for _, stageID := range cacheStagesIDs {
		if stageDesc, err := m.getStageDescription(ctx, stageID, getStageDescriptionOptions{StageShouldExist: true, WithManifestCache: true}); err != nil {
			return false, nil, err
		} else {
			stages = append(stages, stageDesc)
		}
	}

	return cacheExists, stages, err
}

func (m *StagesStorageManager) atomicGetStagesByDigestWithCacheReset(ctx context.Context, stageName, stageSig string) ([]*image.StageDescription, error) {
	if lock, err := m.StorageLockManager.LockStageCache(ctx, m.ProjectName, stageSig); err != nil {
		return nil, fmt.Errorf("error locking project %s stage %s cache: %s", m.ProjectName, stageSig, err)
	} else {
		defer m.StorageLockManager.Unlock(ctx, lock)
	}

	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Info().LogProcess("Get %s stages by digest %s from stages storage", stageName, stageSig).
		DoError(func() error {
			var err error
			stageIDs, err = m.StagesStorage.GetStagesIDsByDigest(ctx, m.ProjectName, stageSig)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from stages storage: %s", m.StagesStorage.String(), stageSig, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	var stages []*image.StageDescription
	for _, stageID := range stageIDs {
		if stageDesc, err := m.getStageDescription(ctx, stageID, getStageDescriptionOptions{StageShouldExist: false, WithManifestCache: true}); err != nil {
			return nil, err
		} else if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
			continue
		} else {
			stages = append(stages, stageDesc)
		}
	}

	if err := logboek.Context(ctx).Info().LogProcess("Storing %s stages by digest %s into stages storage cache", stageName, stageSig).
		DoError(func() error {
			if err := m.StagesStorageCache.StoreStagesByDigest(ctx, m.ProjectName, stageSig, stageIDs); err != nil {
				return fmt.Errorf("error storing stage %s images by digest %s into stages storage cache: %s", stageName, stageSig, err)
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return stages, nil
}

type getStageDescriptionOptions struct {
	StageShouldExist  bool
	WithManifestCache bool
}

func (m *StagesStorageManager) getStageDescription(ctx context.Context, stageID image.StageID, opts getStageDescriptionOptions) (*image.StageDescription, error) {
	stageImageName := m.StagesStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.UniqueID)

	if opts.WithManifestCache {
		logboek.Context(ctx).Debug().LogF("Getting image %s info from manifest cache...\n", stageImageName)
		if imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, m.StagesStorage.String(), stageImageName); err != nil {
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

	logboek.Context(ctx).Debug().LogF("Getting digest %q uniqueID %d stage info from %s...\n", stageID.Digest, stageID.UniqueID, m.StagesStorage.String())
	if stageDesc, err := m.StagesStorage.GetStageDescription(ctx, m.ProjectName, stageID.Digest, stageID.UniqueID); err != nil {
		return nil, fmt.Errorf("error getting digest %q uniqueID %d stage info from %s: %s", stageID.Digest, stageID.UniqueID, m.StagesStorage.String(), err)
	} else if stageDesc != nil {
		logboek.Context(ctx).Debug().LogF("Storing image %s info into manifest cache\n", stageImageName)
		if err := image.CommonManifestCache.StoreImageInfo(ctx, m.StagesStorage.String(), stageDesc.Info); err != nil {
			return nil, fmt.Errorf("error storing image %s info into manifest cache: %s", stageImageName, err)
		}

		return stageDesc, nil
	} else if opts.StageShouldExist {
		logboek.Context(ctx).Warn().LogF("Invalid stage image %q! Stage is no longer available in the %s. Stages storage cache for project %q should be reset!\n", stageImageName, m.StagesStorage.String(), m.ProjectName)
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
