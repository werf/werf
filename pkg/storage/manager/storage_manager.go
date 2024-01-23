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
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/util/parallel"
	"github.com/werf/werf/pkg/werf"
)

var (
	ErrUnexpectedStagesStorageState = errors.New("unexpected stages storage state")
	ErrStageNotFound                = errors.New("stage not found")
)

func IsErrUnexpectedStagesStorageState(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrUnexpectedStagesStorageState.Error())
	}
	return false
}

func IsErrStageNotFound(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrStageNotFound.Error())
	}
	return false
}

type ForEachDeleteStageOptions struct {
	storage.DeleteImageOptions
	storage.FilterStagesAndProcessRelatedDataOptions
}

type StorageOptions struct {
	ContainerBackend container_backend.ContainerBackend
	DockerRegistry   docker_registry.GenericApiInterface
}

type StorageManagerInterface interface {
	InitCache(ctx context.Context) error

	GetStagesStorage() storage.PrimaryStagesStorage
	GetFinalStagesStorage() storage.StagesStorage
	GetSecondaryStagesStorageList() []storage.StagesStorage
	GetCacheStagesStorageList() []storage.StagesStorage

	GetImageInfoGetter(imageName string, desc *image.StageDescription, opts image.InfoGetterOptions) *image.InfoGetter

	EnableParallel(parallelTasksLimit int)
	MaxNumberOfWorkers() int
	GenerateStageUniqueID(digest string, stages []*image.StageDescription) (string, int64)

	GetImageInfo(ctx context.Context, ref string, opts StorageOptions) (*image.Info, error)

	LockStageImage(ctx context.Context, imageName string) error
	GetStagesByDigest(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error)
	GetStagesByDigestWithCache(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error)
	GetStagesByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error)
	GetStagesByDigestFromStagesStorageWithCache(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error)
	GetStageDescriptionList(ctx context.Context) ([]*image.StageDescription, error)
	GetStageDescriptionListWithCache(ctx context.Context) ([]*image.StageDescription, error)
	GetFinalStageDescriptionList(ctx context.Context) ([]*image.StageDescription, error)

	FetchStage(ctx context.Context, containerBackend container_backend.ContainerBackend, stg stage.Interface) error
	SelectSuitableStage(ctx context.Context, c stage.Conveyor, stg stage.Interface, stages []*image.StageDescription) (*image.StageDescription, error)
	CopySuitableByDigestStage(ctx context.Context, stageDesc *image.StageDescription, sourceStagesStorage, destinationStagesStorage storage.StagesStorage, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDescription, error)
	CopyStageIntoCacheStorages(ctx context.Context, stageID image.StageID, cacheStagesStorages []storage.StagesStorage, opts CopyStageIntoStorageOptions) error
	CopyStageIntoFinalStorage(ctx context.Context, stageID image.StageID, finalStagesStorage storage.StagesStorage, opts CopyStageIntoStorageOptions) (*image.StageDescription, error)

	ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error
	ForEachDeleteFinalStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error
	ForEachRmImageMetadata(ctx context.Context, projectName, imageNameOrID string, stageIDCommitList map[string][]string, f func(ctx context.Context, commit, stageID string, err error) error) error
	ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error
	ForEachGetImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.ImportMetadata, err error) error) error
	ForEachRmImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, id string, err error) error) error
	ForEachGetStageCustomTagMetadata(ctx context.Context, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.CustomTagMetadata, err error) error) error
	ForEachDeleteStageCustomTag(ctx context.Context, ids []string, f func(ctx context.Context, tag string, err error) error) error
}

// NOTE: FetchStage is a legacy option, which could in theory be removed.
// NOTE: FetchStage and ContainerBackend options used for a case of copying between local and remote storage.
// NOTE: Remote->Local copy does not need FetchStage and ContainerBackend options.
type CopyStageIntoStorageOptions struct {
	FetchStage           stage.Interface
	ContainerBackend     container_backend.ContainerBackend
	ShouldBeBuiltMode    bool
	IsMultiplatformImage bool
	LogDetailedName      string
}

func RetryOnUnexpectedStagesStorageState(_ context.Context, _ StorageManagerInterface, f func() error) error {
Retry:
	err := f()

	if IsErrUnexpectedStagesStorageState(err) {
		goto Retry
	}

	return err
}

func NewStorageManager(projectName string, stagesStorage storage.PrimaryStagesStorage, finalStagesStorage storage.StagesStorage, secondaryStagesStorageList, cacheStagesStorageList []storage.StagesStorage, storageLockManager storage.LockManager) *StorageManager {
	return &StorageManager{
		ProjectName:        projectName,
		StorageLockManager: storageLockManager,

		StagesStorage:              stagesStorage,
		FinalStagesStorage:         finalStagesStorage,
		CacheStagesStorageList:     cacheStagesStorageList,
		SecondaryStagesStorageList: secondaryStagesStorageList,
	}
}

type StagesList struct {
	Mux      sync.Mutex
	StageIDs []image.StageID
}

func NewStagesList(stageIDs []image.StageID) *StagesList {
	return &StagesList{
		StageIDs: stageIDs,
	}
}

func (stages *StagesList) GetStageIDs() []image.StageID {
	stages.Mux.Lock()
	defer stages.Mux.Unlock()

	var res []image.StageID

	for _, stg := range stages.StageIDs {
		res = append(res, stg)
	}

	return res
}

func (stages *StagesList) AddStageID(stageID image.StageID) {
	stages.Mux.Lock()
	defer stages.Mux.Unlock()

	for _, stg := range stages.StageIDs {
		if stg.IsEqual(stageID) {
			return
		}
	}

	stages.StageIDs = append(stages.StageIDs, stageID)
}

type StorageManager struct {
	parallel           bool
	parallelTasksLimit int

	ProjectName string

	StorageLockManager storage.LockManager

	StagesStorage              storage.PrimaryStagesStorage
	FinalStagesStorage         storage.StagesStorage
	CacheStagesStorageList     []storage.StagesStorage
	SecondaryStagesStorageList []storage.StagesStorage

	// These will be released automatically when current process exits
	SharedHostImagesLocks []lockgate.LockHandle

	FinalStagesListCacheMux sync.Mutex
	FinalStagesListCache    *StagesList
}

func (m *StorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return m.StagesStorage
}

func (m *StorageManager) GetFinalStagesStorage() storage.StagesStorage {
	return m.FinalStagesStorage
}

func (m *StorageManager) GetSecondaryStagesStorageList() []storage.StagesStorage {
	return m.SecondaryStagesStorageList
}

func (m *StorageManager) GetCacheStagesStorageList() []storage.StagesStorage {
	return m.CacheStagesStorageList
}

func (m *StorageManager) GetServiceValuesRepo() string {
	if m.FinalStagesStorage != nil {
		return m.FinalStagesStorage.String()
	}
	return m.StagesStorage.String()
}

func (m *StorageManager) GetImageInfoGetter(imageName string, desc *image.StageDescription, opts image.InfoGetterOptions) *image.InfoGetter {
	if m.FinalStagesStorage != nil {
		finalImageName := m.FinalStagesStorage.ConstructStageImageName(m.ProjectName, desc.StageID.Digest, desc.StageID.UniqueID)
		return image.NewInfoGetter(imageName, finalImageName, opts)
	}

	return image.NewInfoGetter(imageName, desc.Info.Name, opts)
}

func (m *StorageManager) InitCache(ctx context.Context) error {
	logboek.Context(ctx).Info().LogF("Initializing storage manager cache\n")

	if m.FinalStagesStorage != nil {
		if _, err := m.getOrCreateFinalStagesListCache(ctx); err != nil {
			return fmt.Errorf("unable to get or create final stages list cache: %w", err)
		}
	}

	return nil
}

func (m *StorageManager) EnableParallel(parallelTasksLimit int) {
	m.parallel = true
	m.parallelTasksLimit = parallelTasksLimit
}

func (m *StorageManager) MaxNumberOfWorkers() int {
	if m.parallel && m.parallelTasksLimit > 0 {
		return m.parallelTasksLimit
	}

	return 1
}

func (m *StorageManager) GetStageDescriptionListWithCache(ctx context.Context) ([]*image.StageDescription, error) {
	return m.getStageDescriptionList(ctx, storage.WithCache())
}

func (m *StorageManager) GetStageDescriptionList(ctx context.Context) ([]*image.StageDescription, error) {
	return m.getStageDescriptionList(ctx)
}

func (m *StorageManager) getStageDescriptionList(ctx context.Context, opts ...storage.Option) ([]*image.StageDescription, error) {
	stageIDs, err := m.StagesStorage.GetStagesIDs(ctx, m.ProjectName, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting stages ids from %s: %w", m.StagesStorage, err)
	}

	var mutex sync.Mutex
	var stages []*image.StageDescription

	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, m.StagesStorage, m.CacheStagesStorageList, getStageDescriptionOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			return fmt.Errorf("error getting stage %s description: %w", stageID.String(), err)
		}

		if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
			return nil
		}

		mutex.Lock()
		defer mutex.Unlock()
		stages = append(stages, stageDesc)

		return nil
	}); err != nil {
		return nil, err
	}

	return stages, nil
}

func (m *StorageManager) GetFinalStageDescriptionList(ctx context.Context) ([]*image.StageDescription, error) {
	existingStagesListCache, err := m.getOrCreateFinalStagesListCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting existing stages list of final repo %s: %w", m.FinalStagesStorage.String(), err)
	}

	logboek.Context(ctx).Debug().LogF("[%p] Got existing final stages list cache: %#v\n", m, existingStagesListCache.StageIDs)

	stageIDs := existingStagesListCache.GetStageIDs()

	var mutex sync.Mutex
	var stages []*image.StageDescription

	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, m.FinalStagesStorage, nil, getStageDescriptionOptions{WithLocalManifestCache: true})
		if err != nil {
			return fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), m.FinalStagesStorage.String(), err)
		}

		if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.FinalStagesStorage.String())
			return nil
		}

		mutex.Lock()
		defer mutex.Unlock()
		stages = append(stages, stageDesc)

		return nil
	}); err != nil {
		return nil, err
	}

	return stages, nil
}

func (m *StorageManager) ForEachDeleteFinalStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error {
	return parallel.DoTasks(ctx, len(stagesDescriptions), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDescription := stagesDescriptions[taskId]

		err := m.FinalStagesStorage.DeleteStage(ctx, stageDescription, options.DeleteImageOptions)
		return f(ctx, stageDescription, err)
	})
}

func (m *StorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stagesDescriptions []*image.StageDescription, f func(ctx context.Context, stageDesc *image.StageDescription, err error) error) error {
	if localStagesStorage, isLocal := m.StagesStorage.(*storage.LocalStagesStorage); isLocal {
		filteredStagesDescriptions, err := localStagesStorage.FilterStagesAndProcessRelatedData(ctx, stagesDescriptions, options.FilterStagesAndProcessRelatedDataOptions)
		if err != nil {
			return fmt.Errorf("error filtering local docker server stages: %w", err)
		}

		stagesDescriptions = filteredStagesDescriptions
	}

	return parallel.DoTasks(ctx, len(stagesDescriptions), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDescription := stagesDescriptions[taskId]

		for _, cacheStagesStorage := range m.CacheStagesStorageList {
			if err := cacheStagesStorage.DeleteStage(ctx, stageDescription, options.DeleteImageOptions); err != nil {
				logboek.Context(ctx).Warn().LogF("Unable to delete stage %s from the cache stages storage %s: %s\n", stageDescription.StageID.String(), cacheStagesStorage.String(), err)
			}
		}

		err := m.StagesStorage.DeleteStage(ctx, stageDescription, options.DeleteImageOptions)
		return f(ctx, stageDescription, err)
	})
}

func (m *StorageManager) GetImageInfo(ctx context.Context, ref string, opts StorageOptions) (*image.Info, error) {
	info, err := m.getImageInfoFromContainerBackend(ctx, ref, opts.ContainerBackend)
	if err != nil {
		return nil, err
	}
	if info != nil {
		return info, err
	}
	return m.getImageInfoFromRegistry(ctx, ref, opts.DockerRegistry)
}

func (m *StorageManager) getImageInfoFromContainerBackend(ctx context.Context, ref string, containerBackend container_backend.ContainerBackend) (*image.Info, error) {
	return containerBackend.GetImageInfo(ctx, ref, container_backend.GetImageInfoOpts{})
}

func (m *StorageManager) getImageInfoFromRegistry(ctx context.Context, ref string, dockerRegistry docker_registry.GenericApiInterface) (*image.Info, error) {
	cfg, err := dockerRegistry.GetRepoImageConfigFile(ctx, ref)
	if err != nil {
		return nil, err
	}
	return docker_registry.NewImageInfoFromRegistryConfig(ref, cfg), nil
}

func (m *StorageManager) LockStageImage(ctx context.Context, imageName string) error {
	imageLockName := container_backend.ImageLockName(imageName)

	_, lock, err := werf.AcquireHostLock(ctx, imageLockName, lockgate.AcquireOptions{Shared: true})
	if err != nil {
		return fmt.Errorf("error locking %q shared lock: %w", imageLockName, err)
	}

	m.SharedHostImagesLocks = append(m.SharedHostImagesLocks, lock)

	return nil
}

func doFetchStage(ctx context.Context, projectName string, stagesStorage storage.StagesStorage, stageID image.StageID, img container_backend.LegacyImageInterface) error {
	err := logboek.Context(ctx).Info().LogProcess("Check manifest availability").DoError(func() error {
		freshStageDescription, err := stagesStorage.GetStageDescription(ctx, projectName, stageID)
		if err != nil {
			return fmt.Errorf("unable to get stage description: %w", err)
		}

		if freshStageDescription == nil {
			return ErrStageNotFound
		}

		img.SetStageDescription(freshStageDescription)

		return nil
	})
	if err != nil {
		return err
	}

	return logboek.Context(ctx).Info().LogProcess("Fetch image").DoError(func() error {
		logboek.Context(ctx).Debug().LogF("Image name: %s\n", img.Name())

		if err := stagesStorage.FetchImage(ctx, img); err != nil {
			return fmt.Errorf("unable to fetch stage %s image %s: %w", stageID.String(), img.Name(), err)
		}
		return nil
	})
}

func (m *StorageManager) FetchStage(ctx context.Context, containerBackend container_backend.ContainerBackend, stg stage.Interface) error {
	logboek.Context(ctx).Debug().LogF("-- StagesManager.FetchStage %s\n", stg.LogDetailedName())

	if err := m.LockStageImage(ctx, stg.GetStageImage().Image.Name()); err != nil {
		return fmt.Errorf("error locking stage image %q: %w", stg.GetStageImage().Image.Name(), err)
	}

	shouldFetch, err := m.StagesStorage.ShouldFetchImage(ctx, stg.GetStageImage().Image)
	if err != nil {
		return fmt.Errorf("error checking should fetch image: %w", err)
	}
	if !shouldFetch {
		imageName := m.StagesStorage.ConstructStageImageName(m.ProjectName, stg.GetStageImage().Image.GetStageDescription().StageID.Digest, stg.GetStageImage().Image.GetStageDescription().StageID.UniqueID)

		logboek.Context(ctx).Info().LogF("Image %s exists, will not perform fetch\n", imageName)

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, imageName); err != nil {
			return fmt.Errorf("error accessing last recently used images cache for %s: %w", imageName, err)
		}

		return nil
	}

	var fetchedImg container_backend.LegacyImageInterface
	var cacheStagesStorageListToRefill []storage.StagesStorage

	fetchStageFromCache := func(stagesStorage storage.StagesStorage) (container_backend.LegacyImageInterface, error) {
		stageID := stg.GetStageImage().Image.GetStageDescription().StageID
		imageName := stagesStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.UniqueID)
		stageImage := container_backend.NewLegacyStageImage(nil, imageName, containerBackend, stg.GetStageImage().Image.GetTargetPlatform())

		shouldFetch, err := stagesStorage.ShouldFetchImage(ctx, stageImage)
		if err != nil {
			return nil, fmt.Errorf("error checking should fetch image from cache repo %s: %w", stagesStorage.String(), err)
		}

		if shouldFetch {
			logboek.Context(ctx).Info().LogF("Cache repo image %s does not exist locally, will perform fetch\n", stageImage.Name())

			proc := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from %s", stg.LogDetailedName(), stagesStorage.String())
			proc.Start()

			err := doFetchStage(ctx, m.ProjectName, stagesStorage, *stageID, stageImage)

			if IsErrStageNotFound(err) {
				logboek.Context(ctx).Default().LogF("Stage not found\n")
				proc.End()
				return nil, err
			}

			if err != nil {
				proc.Fail()
				return nil, err
			}

			proc.End()

			if err := storeStageDescriptionIntoLocalManifestCache(ctx, m.ProjectName, *stageID, stagesStorage, stageImage.GetStageDescription()); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", imageName, err)
			}
		} else {
			logboek.Context(ctx).Info().LogF("Cache repo image %s exists locally, will not perform fetch\n", stageImage.Name())

			stageDesc, err := getStageDescription(ctx, m.ProjectName, *stageID, stagesStorage, nil, getStageDescriptionOptions{WithLocalManifestCache: true})
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), m.FinalStagesStorage.String(), err)
			}
			if stageDesc == nil {
				return nil, ErrStageNotFound
			}
			stageImage.SetStageDescription(stageDesc)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, stageImage.Name()); err != nil {
			return nil, fmt.Errorf("error accessing last recently used images cache for %s: %w", stageImage.Name(), err)
		}

		return stageImage, nil
	}

	prepareCacheStageAsPrimary := func(cacheImg container_backend.LegacyImageInterface, primaryStage stage.Interface) error {
		stageID := primaryStage.GetStageImage().Image.GetStageDescription().StageID
		primaryImageName := m.StagesStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.UniqueID)

		// TODO(buildah): check no bugs introduced by removing of following calls
		// if err := containerBackend.RenameImage(ctx, cacheDockerImage, primaryImageName, false); err != nil {
		//	 return fmt.Errorf("unable to rename image %s to %s: %w", fetchedDockerImage.Image.Name(), primaryImageName, err)
		// }

		// if err := containerBackend.RefreshImageObject(ctx, &container_backend.Image{Image: primaryStage.GetImage()}); err != nil {
		//	 return fmt.Errorf("unable to refresh stage image %s: %w", primaryStage.GetImage().Name(), err)
		// }

		// TODO(buildah): check no bugs introduced by removing of following calls
		// if err := storeStageDescriptionIntoLocalManifestCache(ctx, m.ProjectName, *stageID, m.StagesStorage, ConvertStageDescriptionForStagesStorage(cacheDockerImage.Image.GetStageDescription(), m.StagesStorage)); err != nil {
		if err := storeStageDescriptionIntoLocalManifestCache(ctx, m.ProjectName, *stageID, m.StagesStorage, cacheImg.GetStageDescription()); err != nil {
			return fmt.Errorf("error storing stage %s description into local manifest cache: %w", primaryImageName, err)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, primaryImageName); err != nil {
			return fmt.Errorf("error accessing last recently used images cache for %s: %w", primaryImageName, err)
		}

		return nil
	}

	for _, cacheStagesStorage := range m.CacheStagesStorageList {
		cacheImg, err := fetchStageFromCache(cacheStagesStorage)
		if err != nil {
			if !IsErrStageNotFound(err) {
				logboek.Context(ctx).Warn().LogF("Unable to fetch stage %s from cache stages storage %s: %s\n", stg.GetStageImage().Image.GetStageDescription().StageID.String(), cacheStagesStorage.String(), err)
			}

			cacheStagesStorageListToRefill = append(cacheStagesStorageListToRefill, cacheStagesStorage)

			continue
		}

		if err := prepareCacheStageAsPrimary(cacheImg, stg); err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to prepare stage %s fetched from cache stages storage %s as a primary: %s\n", cacheImg.Name(), cacheStagesStorage.String(), err)

			cacheStagesStorageListToRefill = append(cacheStagesStorageListToRefill, cacheStagesStorage)

			continue
		}

		fetchedImg = cacheImg
		break
	}

	if fetchedImg == nil {
		stageID := stg.GetStageImage().Image.GetStageDescription().StageID
		img := stg.GetStageImage()

		err := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from %s", stg.LogDetailedName(), m.StagesStorage.String()).
			DoError(func() error {
				return doFetchStage(ctx, m.ProjectName, m.StagesStorage, *stageID, img.Image)
			})

		if IsErrStageNotFound(err) {
			logboek.Context(ctx).Error().LogF("Stage %s image %s is no longer available!\n", stg.LogDetailedName(), stg.GetStageImage().Image.Name())
			return ErrUnexpectedStagesStorageState
		}

		if storage.IsErrBrokenImage(err) {
			logboek.Context(ctx).Error().LogF("Broken stage %s image %s!\n", stg.LogDetailedName(), stg.GetStageImage().Image.Name())

			logboek.Context(ctx).Error().LogF("Will mark image %s as rejected in the stages storage %s\n", stg.GetStageImage().Image.Name(), m.StagesStorage.String())
			if err := m.StagesStorage.RejectStage(ctx, m.ProjectName, stageID.Digest, stageID.UniqueID); err != nil {
				return fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %w", stg.LogDetailedName(), stg.GetStageImage().Image.Name(), m.StagesStorage.String(), err)
			}

			return ErrUnexpectedStagesStorageState
		}

		if err != nil {
			return fmt.Errorf("unable to fetch stage %s from stages storage %s: %w", stageID.String(), m.StagesStorage.String(), err)
		}

		fetchedImg = img.Image
	}

	for _, cacheStagesStorage := range cacheStagesStorageListToRefill {
		stageID := stg.GetStageImage().Image.GetStageDescription().StageID

		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", stg.LogDetailedName(), cacheStagesStorage.String()).
			DoError(func() error {
				if _, err := m.CopyStage(ctx, m.StagesStorage, cacheStagesStorage, *stageID, CopyStageOptions{
					ContainerBackend: containerBackend,
					LegacyImage:      fetchedImg,
				}); err != nil {
					return fmt.Errorf("unable to copy stage %s into cache stages storage %s: %w", stageID.String(), cacheStagesStorage.String(), err)
				}
				return nil
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Warning %s\n", err)
		}
	}

	return nil
}

func (m *StorageManager) CopyStageIntoCacheStorages(ctx context.Context, stageID image.StageID, cacheStagesStorageList []storage.StagesStorage, opts CopyStageIntoStorageOptions) error {
	for _, cache := range cacheStagesStorageList {
		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", opts.LogDetailedName, cache.String()).
			DoError(func() error {
				copyOpts := CopyStageOptions{ContainerBackend: opts.ContainerBackend}
				if opts.FetchStage != nil {
					copyOpts.FetchStage = opts.FetchStage
					copyOpts.LegacyImage = opts.FetchStage.GetStageImage().Image
				}
				if _, err := m.CopyStage(ctx, m.StagesStorage, cache, stageID, copyOpts); err != nil {
					return fmt.Errorf("unable to copy stage %s into cache stages storage %s: %w", stageID.String(), cache.String(), err)
				}
				return nil
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Warning: %s\n", err)
		}
	}
	return nil
}

func (m *StorageManager) getOrCreateFinalStagesListCache(ctx context.Context) (*StagesList, error) {
	m.FinalStagesListCacheMux.Lock()
	defer m.FinalStagesListCacheMux.Unlock()

	if m.FinalStagesListCache != nil {
		return m.FinalStagesListCache, nil
	}

	stageIDs, err := m.FinalStagesStorage.GetStagesIDs(ctx, m.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("unable to get final repo stages list: %w", err)
	}
	m.FinalStagesListCache = NewStagesList(stageIDs)

	return m.FinalStagesListCache, nil
}

func (m *StorageManager) CopyStageIntoFinalStorage(ctx context.Context, stageID image.StageID, finalStagesStorage storage.StagesStorage, opts CopyStageIntoStorageOptions) (*image.StageDescription, error) {
	existingStagesListCache, err := m.getOrCreateFinalStagesListCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting existing stages list of final repo %s: %w", finalStagesStorage.String(), err)
	}

	logboek.Context(ctx).Debug().LogF("[%p] Got existing final stages list cache: %#v\n", m, existingStagesListCache.StageIDs)

	finalImageName := finalStagesStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.UniqueID)

	for _, existingStg := range existingStagesListCache.GetStageIDs() {
		if existingStg.IsEqual(stageID) {
			desc, err := m.GetFinalStagesStorage().GetStageDescription(ctx, m.ProjectName, stageID)
			if err != nil {
				return nil, fmt.Errorf("unable to get stage %s descriptor from final repo %s: %w", stageID.String(), m.GetFinalStagesStorage().String(), err)
			}
			if desc != nil {
				logboek.Context(ctx).Info().LogF("Stage %s already exists in the final repo, skipping\n", stageID.String())

				logboek.Context(ctx).Default().LogFHighlight("Use previously built final image for %s\n", opts.LogDetailedName)
				container_backend.LogImageName(ctx, finalImageName)

				return desc, nil
			}
		}
	}

	if opts.ShouldBeBuiltMode {
		return nil, fmt.Errorf("%s with digest %s is not exist in the final repo", opts.LogDetailedName, stageID.Digest)
	}

	var desc *image.StageDescription

	err = logboek.Context(ctx).Default().LogProcess("Copy stage %s into the final repo", opts.LogDetailedName).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			copyOpts := CopyStageOptions{
				ContainerBackend:     opts.ContainerBackend,
				IsMultiplatformImage: opts.IsMultiplatformImage,
			}
			if opts.FetchStage != nil {
				copyOpts.FetchStage = opts.FetchStage
				copyOpts.LegacyImage = opts.FetchStage.GetStageImage().Image
			}
			desc, err = m.CopyStage(ctx, m.StagesStorage, finalStagesStorage, stageID, copyOpts)
			if err != nil {
				return fmt.Errorf("unable to copy stage %s into the final repo %s: %w", stageID.String(), finalStagesStorage.String(), err)
			}

			logboek.Context(ctx).Default().LogFDetails("  name: %s\n", finalImageName)

			return nil
		})
	if err != nil {
		return nil, err
	}

	existingStagesListCache.AddStageID(stageID)
	logboek.Context(ctx).Debug().LogF("Updated existing final stages list: %#v\n", m.FinalStagesListCache.StageIDs)

	return desc, nil
}

func (m *StorageManager) SelectSuitableStage(ctx context.Context, c stage.Conveyor, stg stage.Interface, stages []*image.StageDescription) (*image.StageDescription, error) {
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

func (m *StorageManager) GetStagesByDigestWithCache(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error) {
	return m.GetStagesByDigestFromStagesStorageWithCache(ctx, stageName, stageDigest, m.StagesStorage)
}

func (m *StorageManager) GetStagesByDigest(ctx context.Context, stageName, stageDigest string) ([]*image.StageDescription, error) {
	return m.GetStagesByDigestFromStagesStorage(ctx, stageName, stageDigest, m.StagesStorage)
}

func (m *StorageManager) GetStagesByDigestFromStagesStorageWithCache(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error) {
	cachedStageDescriptionList, err := m.getStagesByDigestFromStagesStorage(ctx, stageName, stageDigest, stagesStorage, storage.WithCache())
	if err != nil {
		return nil, err
	}

	if len(cachedStageDescriptionList) != 0 {
		return cachedStageDescriptionList, nil
	}

	return m.getStagesByDigestFromStagesStorage(ctx, stageName, stageDigest, stagesStorage)
}

func (m *StorageManager) GetStagesByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage) ([]*image.StageDescription, error) {
	return m.getStagesByDigestFromStagesStorage(ctx, stageName, stageDigest, stagesStorage)
}

func (m *StorageManager) getStagesByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage, opts ...storage.Option) ([]*image.StageDescription, error) {
	stageIDs, err := m.getStagesIDsByDigestFromStagesStorage(ctx, stageName, stageDigest, stagesStorage, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get stages ids from %s by digest %s for stage %s: %w", stagesStorage.String(), stageDigest, stageName, err)
	}

	stages, err := m.getStagesDescriptions(ctx, stageIDs, stagesStorage, m.CacheStagesStorageList)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage descriptions by ids from %s: %w", stagesStorage.String(), err)
	}

	return stages, nil
}

func (m *StorageManager) CopySuitableByDigestStage(ctx context.Context, stageDesc *image.StageDescription, sourceStagesStorage, destinationStagesStorage storage.StagesStorage, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDescription, error) {
	img := container_backend.NewLegacyStageImage(nil, stageDesc.Info.Name, containerBackend, targetPlatform)

	logboek.Context(ctx).Info().LogF("Fetching %s\n", img.Name())
	if err := sourceStagesStorage.FetchImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to fetch %s from %s: %w", stageDesc.Info.Name, sourceStagesStorage.String(), err)
	}

	newImageName := destinationStagesStorage.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.UniqueID)
	logboek.Context(ctx).Info().LogF("Renaming image %s to %s\n", img.Name(), newImageName)
	if err := containerBackend.RenameImage(ctx, img, newImageName, false); err != nil {
		return nil, err
	}

	logboek.Context(ctx).Info().LogF("Storing %s\n", newImageName)
	if err := destinationStagesStorage.StoreImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to store %s to %s: %w", stageDesc.Info.Name, destinationStagesStorage.String(), err)
	}

	if destinationStageDesc, err := getStageDescription(ctx, m.ProjectName, *stageDesc.StageID, destinationStagesStorage, m.CacheStagesStorageList, getStageDescriptionOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()}); err != nil {
		return nil, fmt.Errorf("unable to get stage %s description from %s: %w", stageDesc.StageID.String(), destinationStagesStorage.String(), err)
	} else {
		return destinationStageDesc, nil
	}
}

func (m *StorageManager) getWithLocalManifestCacheOption() bool {
	return m.StagesStorage.Address() != storage.LocalStorageAddress
}

func (m *StorageManager) getStagesIDsByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, stagesStorage storage.StagesStorage, opts ...storage.Option) ([]image.StageID, error) {
	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Info().LogProcess("Get %s stages by digest %s from storage", stageName, stageDigest).
		DoError(func() error {
			var err error
			stageIDs, err = stagesStorage.GetStagesIDsByDigest(ctx, m.ProjectName, stageDigest, opts...)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from storage: %w", m.StagesStorage.String(), stageDigest, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	return stageIDs, nil
}

func (m *StorageManager) getStagesDescriptions(ctx context.Context, stageIDs []image.StageID, stagesStorage storage.StagesStorage, cacheStagesStorageList []storage.StagesStorage) ([]*image.StageDescription, error) {
	var stages []*image.StageDescription
	for _, stageID := range stageIDs {
		stageDesc, err := getStageDescription(ctx, m.ProjectName, stageID, stagesStorage, cacheStagesStorageList, getStageDescriptionOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			return nil, err
		}

		if stageDesc == nil {
			logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s\n", stageID.String(), m.StagesStorage.String())
			continue
		}

		stages = append(stages, stageDesc)
	}

	return stages, nil
}

type getStageDescriptionOptions struct {
	WithLocalManifestCache bool
}

func getStageDescriptionFromLocalManifestCache(ctx context.Context, projectName string, stageID image.StageID, stagesStorage storage.StagesStorage) (*image.StageDescription, error) {
	stageImageName := stagesStorage.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)

	logboek.Context(ctx).Debug().LogF("Getting image %s info from the manifest cache...\n", stageImageName)
	imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, stagesStorage.String(), stageImageName)
	if err != nil {
		return nil, fmt.Errorf("error getting image %s info: %w", stageImageName, err)
	}

	if imgInfo != nil {
		logboek.Context(ctx).Info().LogF("Got image %s info from the manifest cache (CACHE HIT)\n", stageImageName)

		return &image.StageDescription{
			StageID: image.NewStageID(stageID.Digest, stageID.UniqueID),
			Info:    imgInfo,
		}, nil
	} else {
		logboek.Context(ctx).Info().LogF("Not found %s image info in the manifest cache (CACHE MISS)\n", stageImageName)
	}

	return nil, nil
}

func ConvertStageDescriptionForStagesStorage(stageDesc *image.StageDescription, stagesStorage storage.StagesStorage) *image.StageDescription {
	return &image.StageDescription{
		StageID: image.NewStageID(stageDesc.StageID.Digest, stageDesc.StageID.UniqueID),
		Info: &image.Info{
			Name:              fmt.Sprintf("%s:%s-%d", stagesStorage.Address(), stageDesc.StageID.Digest, stageDesc.StageID.UniqueID),
			Repository:        stagesStorage.Address(),
			Tag:               stageDesc.Info.Tag,
			RepoDigest:        stageDesc.Info.RepoDigest,
			ID:                stageDesc.Info.ID,
			ParentID:          stageDesc.Info.ParentID,
			Labels:            stageDesc.Info.Labels,
			Size:              stageDesc.Info.Size,
			CreatedAtUnixNano: stageDesc.Info.CreatedAtUnixNano,
			OnBuild:           stageDesc.Info.OnBuild,
			Env:               stageDesc.Info.Env,
		},
	}
}

func getStageDescription(ctx context.Context, projectName string, stageID image.StageID, stagesStorage storage.StagesStorage, cacheStagesStorageList []storage.StagesStorage, opts getStageDescriptionOptions) (*image.StageDescription, error) {
	if opts.WithLocalManifestCache {
		stageDesc, err := getStageDescriptionFromLocalManifestCache(ctx, projectName, stageID, stagesStorage)
		if err != nil {
			return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), stagesStorage.String(), err)
		}
		if stageDesc != nil {
			return stageDesc, nil
		}
	}

	for _, cacheStagesStorage := range cacheStagesStorageList {
		if opts.WithLocalManifestCache {
			stageDesc, err := getStageDescriptionFromLocalManifestCache(ctx, projectName, stageID, cacheStagesStorage)
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from the local manifest cache: %w", stageID.String(), err)
			}
			if stageDesc != nil {
				return ConvertStageDescriptionForStagesStorage(stageDesc, stagesStorage), nil
			}
		}

		var stageDesc *image.StageDescription
		err := logboek.Context(ctx).Info().LogProcess("Get stage %s description from cache stages storage %s", stageID.String(), cacheStagesStorage.String()).
			DoError(func() error {
				var err error
				stageDesc, err = cacheStagesStorage.GetStageDescription(ctx, projectName, stageID)

				logboek.Context(ctx).Debug().LogF("Got stage description: %#v\n", stageDesc)
				return err
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to get stage description from cache stages storage %s: %s\n", cacheStagesStorage.String(), err)
			continue
		}

		if stageDesc != nil {
			if opts.WithLocalManifestCache {
				if err := storeStageDescriptionIntoLocalManifestCache(ctx, projectName, stageID, cacheStagesStorage, stageDesc); err != nil {
					return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
				}
			}

			return ConvertStageDescriptionForStagesStorage(stageDesc, stagesStorage), nil
		}
	}

	logboek.Context(ctx).Debug().LogF("Getting digest %q uniqueID %d stage info from %s...\n", stageID.Digest, stageID.UniqueID, stagesStorage.String())
	stageDesc, err := stagesStorage.GetStageDescription(ctx, projectName, stageID)
	switch {
	case storage.IsErrBrokenImage(err):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error getting digest %q uniqueID %d stage info from %s: %w", stageID.Digest, stageID.UniqueID, stagesStorage.String(), err)
	case stageDesc != nil:
		if opts.WithLocalManifestCache {
			if err := storeStageDescriptionIntoLocalManifestCache(ctx, projectName, stageID, stagesStorage, stageDesc); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
			}
		}
		return stageDesc, nil
	default:
		return nil, nil
	}
}

func (m *StorageManager) GenerateStageUniqueID(digest string, stages []*image.StageDescription) (string, int64) {
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

func (m *StorageManager) ForEachRmImageMetadata(ctx context.Context, projectName, imageNameOrID string, stageIDCommitList map[string][]string, f func(ctx context.Context, commit, stageID string, err error) error) error {
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

func (m *StorageManager) ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error {
	return parallel.DoTasks(ctx, len(managedImages), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		managedImage := managedImages[taskId]
		err := m.StagesStorage.RmManagedImage(ctx, projectName, managedImage)
		return f(ctx, managedImage, err)
	})
}

func (m *StorageManager) ForEachGetImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.ImportMetadata, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]
		metadata, err := m.StagesStorage.GetImportMetadata(ctx, projectName, id)
		return f(ctx, id, metadata, err)
	})
}

func (m *StorageManager) ForEachRmImportMetadata(ctx context.Context, projectName string, ids []string, f func(ctx context.Context, id string, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]
		err := m.StagesStorage.RmImportMetadata(ctx, projectName, id)
		return f(ctx, id, err)
	})
}

func (m *StorageManager) ForEachDeleteStageCustomTag(ctx context.Context, ids []string, f func(ctx context.Context, tag string, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]

		if err := m.StagesStorage.DeleteStageCustomTag(ctx, id); err != nil {
			return f(ctx, id, fmt.Errorf("unable to delete stage custom tag: %w", err))
		}
		if err := m.StagesStorage.UnregisterStageCustomTag(ctx, id); err != nil {
			return f(ctx, id, fmt.Errorf("unable to unregister stage custom tag: %w", err))
		}

		return f(ctx, id, nil)
	})
}

func (m *StorageManager) ForEachGetStageCustomTagMetadata(ctx context.Context, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.CustomTagMetadata, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]
		metadata, err := m.StagesStorage.GetStageCustomTagMetadata(ctx, id)
		return f(ctx, id, metadata, err)
	})
}
