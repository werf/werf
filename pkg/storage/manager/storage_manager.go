package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
	"gopkg.in/yaml.v2"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var ErrUnexpectedStorageState = errors.New("unexpected stages storage state")

const maxRetryAttemptsOnUnexpectedStorageState = 4

func IsErrUnexpectedStorageState(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrUnexpectedStorageState.Error())
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

	GetMetaStorage() storage.MetaStorage
	GetImagesRepoStorage() storage.ImagesRepoStorage
	GetImagesRepoStorages() []storage.ImagesRepoStorage
	GetCacheReaders() []storage.StageReader
	GetCacheWriters() []storage.StageWriter

	GetImageInfoGetter(imageName string, desc *image.StageDesc, opts image.InfoGetterOptions) *image.InfoGetter

	EnableParallel(parallelTasksLimit int)
	MaxNumberOfWorkers() int
	GenerateStageDescCreationTs(digest string, stageDescSet image.StageDescSet) (string, int64)

	LockStageImage(ctx context.Context, imageName string) error
	GetStageDescSetByDigest(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error)
	GetStageDescSetByDigestWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error)
	GetStageDescSetByDigestFromReader(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader) (image.StageDescSet, error)
	GetStageDescSetByDigestFromReaderWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader) (image.StageDescSet, error)
	GetStageDescSet(ctx context.Context) (image.StageDescSet, error)
	GetStageDescSetWithCache(ctx context.Context) (image.StageDescSet, error)
	GetImagesRepoStageDescSet(ctx context.Context) (image.StageDescSet, error)

	FetchStage(ctx context.Context, containerBackend container_backend.ContainerBackend, stg stage.Interface) (FetchStageInfo, error)
	SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, stageDescSet image.StageDescSet) (*image.StageDesc, error)
	CopySuitableStageDescByDigest(ctx context.Context, stageDesc *image.StageDesc, src storage.StageReader, dest storage.StageWriter, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDesc, error)
	CopyStageIntoCacheToStorages(ctx context.Context, stageID image.StageID, cacheToStorages []storage.StageWriter, opts CopyStageIntoStorageOptions) error
	CopyStageIntoImagesRepoStorage(ctx context.Context, stageID image.StageID, imagesRepoStorage storage.ImagesRepoStorage, opts CopyStageIntoStorageOptions) (*image.StageDesc, error)

	ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error
	ForEachDeleteImagesRepoStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error
	ForEachRmImageMetadata(ctx context.Context, projectName, imageNameOrID string, stageIDCommitList map[string][]string, f func(ctx context.Context, commit, stageID string, err error) error) error
	ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error
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

func RetryOnUnexpectedStorageState(ctx context.Context, _ StorageManagerInterface, f func() error) error {
	var attempt int
	op := func() (bool, error) {
		attempt++

		if err := f(); err != nil {
			if !IsErrUnexpectedStorageState(err) {
				return false, backoff.Permanent(err)
			}
			return false, fmt.Errorf("exhausted %d retries on unexpected storage state: %w", maxRetryAttemptsOnUnexpectedStorageState-1, err)
		}

		return false, nil
	}

	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("Retrying due to unexpected storage state (attempt %d/%d) in %0.2f seconds ...\n", attempt, maxRetryAttemptsOnUnexpectedStorageState-1, duration.Seconds())
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = 2 * time.Second
	eb.MaxInterval = 10 * time.Second

	_, err := backoff.Retry(ctx, op,
		backoff.WithBackOff(eb),
		backoff.WithMaxTries(maxRetryAttemptsOnUnexpectedStorageState),
		backoff.WithNotify(notify),
	)
	return err
}

func NewStorageManager(projectName string, imagesRepoStorages []storage.ImagesRepoStorage, metaStorage storage.MetaStorage, cacheReaders []storage.StageReader, cacheWriters []storage.StageWriter, storageLockManager lock_manager.Interface) *StorageManager {
	if len(cacheReaders) == 0 {
		panic("cache readers should not be empty")
	}

	return &StorageManager{
		ProjectName:        projectName,
		StorageLockManager: storageLockManager,

		MetaStorage:        metaStorage,
		ImagesRepoStorages: imagesRepoStorages,
		CacheWriters:       cacheWriters,
		CacheReaders:       cacheReaders,
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

	StorageLockManager lock_manager.Interface

	ImagesRepoStorages []storage.ImagesRepoStorage
	MetaStorage        storage.MetaStorage
	CacheWriters       []storage.StageWriter
	CacheReaders       []storage.StageReader

	// These will be released automatically when current process exits
	SharedHostImagesLocks []lockgate.LockHandle

	ImagesRepoStagesListCacheMux sync.Mutex
	ImagesRepoStagesListCache    map[string]*StagesList
}

func (m *StorageManager) GetMetaStorage() storage.MetaStorage {
	return m.MetaStorage
}

func (m *StorageManager) GetImagesRepoStorage() storage.ImagesRepoStorage {
	if len(m.ImagesRepoStorages) > 0 {
		return m.ImagesRepoStorages[0]
	}
	return nil
}

func (m *StorageManager) GetImagesRepoStorages() []storage.ImagesRepoStorage {
	return m.ImagesRepoStorages
}

func (m *StorageManager) GetCacheReaders() []storage.StageReader {
	return m.CacheReaders
}

func (m *StorageManager) GetCacheWriters() []storage.StageWriter {
	return m.CacheWriters
}

func (m *StorageManager) primaryReader() storage.StageReader {
	if len(m.CacheReaders) == 0 {
		panic("cache readers should not be empty")
	}

	return m.CacheReaders[0]
}

func sameStageStorage(a, b storage.BaseStorage) bool {
	return a.String() == b.String() && a.Address() == b.Address()
}

func (m *StorageManager) otherReaders(reader storage.StageReader) []storage.StageReader {
	var readers []storage.StageReader

	for _, r := range m.CacheReaders {
		if sameStageStorage(r, reader) {
			continue
		}

		readers = append(readers, r)
	}

	return readers
}

func (m *StorageManager) GetServiceValuesRepo() string {
	if img := m.GetImagesRepoStorage(); img != nil {
		return img.String()
	}
	if m.MetaStorage != nil {
		return m.MetaStorage.String()
	}
	return ""
}

func (m *StorageManager) GetImageInfoGetter(imageName string, stageDesc *image.StageDesc, opts image.InfoGetterOptions) *image.InfoGetter {
	if img := m.GetImagesRepoStorage(); img != nil {
		finalImageName := img.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.CreationTs)
		return image.NewInfoGetter(imageName, finalImageName, stageDesc.Info.GetDigest(), opts)
	}

	return image.NewInfoGetter(imageName, stageDesc.Info.Name, stageDesc.Info.GetDigest(), opts)
}

func (m *StorageManager) InitCache(ctx context.Context) error {
	logboek.Context(ctx).Info().LogF("Initializing storage manager cache\n")

	for _, imgRepo := range m.ImagesRepoStorages {
		if _, err := m.getOrCreateImagesRepoStagesListCache(ctx, imgRepo); err != nil {
			return fmt.Errorf("unable to get or create images repo stages list cache for %s: %w", imgRepo.String(), err)
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

func (m *StorageManager) GetStageDescSetWithCache(ctx context.Context) (image.StageDescSet, error) {
	return m.getStageDescSet(ctx, storage.WithCache())
}

func (m *StorageManager) GetStageDescSet(ctx context.Context) (image.StageDescSet, error) {
	return m.getStageDescSet(ctx)
}

func (m *StorageManager) getStageDescSet(ctx context.Context, opts ...storage.Option) (image.StageDescSet, error) {
	stageIDsByStorage := map[string][]image.StageID{}
	storageByName := map[string]storage.StageReader{}
	orderedStorageNames := []string{}

	for _, reader := range m.CacheReaders {
		stageIDs, err := reader.GetStagesIDs(ctx, m.ProjectName, opts...)
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to get stages ids from %s: %s\n", reader.String(), err)
			continue
		}

		storageName := reader.String()
		if _, hasKey := stageIDsByStorage[storageName]; !hasKey {
			orderedStorageNames = append(orderedStorageNames, storageName)
		}

		storageByName[storageName] = reader
		stageIDsByStorage[storageName] = append(stageIDsByStorage[storageName], stageIDs...)
	}

	stageIDsByName := map[string]image.StageID{}
	orderedStageIDs := []image.StageID{}
	for _, storageName := range orderedStorageNames {
		for _, stageID := range stageIDsByStorage[storageName] {
			key := stageID.String()
			if _, hasKey := stageIDsByName[key]; hasKey {
				continue
			}

			stageIDsByName[key] = stageID
			orderedStageIDs = append(orderedStageIDs, stageID)
		}
	}

	stageDescSet := image.NewStageDescSet()
	if err := parallel.DoTasks(ctx, len(orderedStageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := orderedStageIDs[taskId]

		var reader storage.StageReader
		for _, storageName := range orderedStorageNames {
			for _, stageStorageID := range stageIDsByStorage[storageName] {
				if !stageStorageID.IsEqual(stageID) {
					continue
				}

				reader = storageByName[storageName]
				break
			}

			if reader != nil {
				break
			}
		}
		if reader == nil {
			return nil
		}

		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, reader, m.otherReaders(reader), getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), reader.String(), err)
				return nil
			}

			return fmt.Errorf("error getting stage %s description: %w", stageID.String(), err)
		}

		stageDescSet.Add(stageDesc)

		return nil
	}); err != nil {
		return nil, err
	}

	return stageDescSet, nil
}

func (m *StorageManager) GetImagesRepoStageDescSet(ctx context.Context) (image.StageDescSet, error) {
	if len(m.ImagesRepoStorages) == 0 {
		return image.NewStageDescSet(), nil
	}

	// Use the primary (first) images repo for the stage desc set
	primaryImagesRepo := m.ImagesRepoStorages[0]
	existingStagesListCache, err := m.getOrCreateImagesRepoStagesListCache(ctx, primaryImagesRepo)
	if err != nil {
		return nil, fmt.Errorf("error getting existing stages list of images repo %s: %w", primaryImagesRepo.String(), err)
	}

	logboek.Context(ctx).Debug().LogF("[%p] Got existing images repo stages list cache: %#v\n", m, existingStagesListCache.StageIDs)

	stageIDs := existingStagesListCache.GetStageIDs()
	stageDescSet := image.NewStageDescSet()
	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, primaryImagesRepo, nil, getStageDescOptions{WithLocalManifestCache: true})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), primaryImagesRepo.String(), err)
				return nil
			}

			return fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), primaryImagesRepo.String(), err)
		}

		stageDescSet.Add(stageDesc)

		return nil
	}); err != nil {
		return nil, err
	}

	return stageDescSet, nil
}

func (m *StorageManager) ForEachDeleteImagesRepoStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error {
	stageDescSet = stageDescSet.Clone()
	return parallel.DoTasks(ctx, stageDescSet.Cardinality(), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDesc, _ := stageDescSet.Pop()

		// Delete from all images repos
		var lastErr error
		for _, imgRepo := range m.ImagesRepoStorages {
			if err := imgRepo.DeleteStage(ctx, stageDesc, options.DeleteImageOptions); err != nil {
				if !storage.IsErrStageUnavailable(err) {
					lastErr = err
				}
			}
		}

		return f(ctx, stageDesc, lastErr)
	})
}

func (m *StorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error {
	for _, cacheReader := range m.CacheReaders {
		localStagesStorage, isLocal := cacheReader.(*storage.LocalStagesStorage)
		if !isLocal {
			continue
		}

		filteredStageDescSet, err := localStagesStorage.FilterStageDescSetAndProcessRelatedData(ctx, stageDescSet, options.FilterStagesAndProcessRelatedDataOptions)
		if err != nil {
			return fmt.Errorf("error filtering local docker server stages: %w", err)
		}

		stageDescSet = filteredStageDescSet
		break
	}

	stageDescSet = stageDescSet.Clone()
	return parallel.DoTasks(ctx, stageDescSet.Cardinality(), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDesc, _ := stageDescSet.Pop()

		for _, cacheWriter := range m.CacheWriters {
			if err := cacheWriter.DeleteStage(ctx, stageDesc, options.DeleteImageOptions); err != nil {
				logboek.Context(ctx).Warn().LogF("Unable to delete stage %s from the cache to storage %s: %s\n", stageDesc.StageID.String(), cacheWriter.String(), err)
			}
		}

		return f(ctx, stageDesc, nil)
	})
}

func (m *StorageManager) LockStageImage(ctx context.Context, imageName string) error {
	imageLockName := container_backend.ImageLockName(imageName)

	_, l, err := werf.HostLocker().AcquireLock(ctx, imageLockName, lockgate.AcquireOptions{Shared: true})
	if err != nil {
		return fmt.Errorf("error locking %q shared lock: %w", imageLockName, err)
	}

	m.SharedHostImagesLocks = append(m.SharedHostImagesLocks, l)

	return nil
}

func doFetchStage(ctx context.Context, projectName string, reader storage.StageReader, stageID image.StageID, img container_backend.LegacyImageInterface) error {
	err := logboek.Context(ctx).Info().LogProcess("Check manifest availability").DoError(func() error {
		freshStageDesc, err := reader.GetStageDesc(ctx, projectName, stageID)
		if err != nil {
			return fmt.Errorf("unable to get stage description: %w", err)
		}

		img.SetStageDesc(freshStageDesc)

		return nil
	})
	if err != nil {
		return err
	}

	return logboek.Context(ctx).Info().LogProcess("Fetch image").DoError(func() error {
		logboek.Context(ctx).Debug().LogF("Image name: %s\n", img.Name())

		if err := reader.FetchImage(ctx, img); err != nil {
			return fmt.Errorf("unable to fetch stage %s image %s: %w", stageID.String(), img.Name(), err)
		}
		return nil
	})
}

const (
	// could not be imported form build
	BaseImageSourceTypeCacheRepo = "cache-repo"
	BaseImageSourceTypeRepo      = "repo"
)

type FetchStageInfo struct {
	BaseImagePulled bool
	BaseImageSource string
}

func (m *StorageManager) FetchStage(ctx context.Context, containerBackend container_backend.ContainerBackend, stg stage.Interface) (FetchStageInfo, error) {
	logboek.Context(ctx).Debug().LogF("-- StagesManager.FetchStage %s\n", stg.LogDetailedName())

	if err := m.LockStageImage(ctx, stg.GetStageImage().Image.Name()); err != nil {
		return FetchStageInfo{}, fmt.Errorf("error locking stage image %q: %w", stg.GetStageImage().Image.Name(), err)
	}

	shouldFetch, err := m.primaryReader().ShouldFetchImage(ctx, stg.GetStageImage().Image)
	if err != nil {
		return FetchStageInfo{}, fmt.Errorf("error checking should fetch image: %w", err)
	}
	if !shouldFetch {
		imageName := m.primaryReader().ConstructStageImageName(m.ProjectName, stg.GetStageImage().Image.GetStageDesc().StageID.Digest, stg.GetStageImage().Image.GetStageDesc().StageID.CreationTs)

		logboek.Context(ctx).Info().LogF("Image %s exists, will not perform fetch\n", imageName)

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, imageName); err != nil {
			return FetchStageInfo{}, fmt.Errorf("error accessing last recently used images cache for %s: %w", imageName, err)
		}

		return FetchStageInfo{BaseImagePulled: false}, nil
	}

	var fetchedImg container_backend.LegacyImageInterface
	var writersToRefill []storage.StageWriter
	var pulled bool
	var source string
	var fetchedReader storage.StageReader
	allStagesUnavailable := true
	var unavailableReader storage.StageReader

	fetchStageFromCache := func(reader storage.StageReader) (container_backend.LegacyImageInterface, error) {
		stageID := stg.GetStageImage().Image.GetStageDesc().StageID
		imageName := reader.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)
		stageImage := container_backend.NewLegacyStageImage(nil, imageName, containerBackend, stg.GetStageImage().Image.GetTargetPlatform())

		shouldFetch, err := reader.ShouldFetchImage(ctx, stageImage)
		if err != nil {
			return nil, fmt.Errorf("error checking should fetch image from cache repo %s: %w", reader.String(), err)
		}

		if shouldFetch {
			logboek.Context(ctx).Info().LogF("Cache repo image %s does not exist locally, will perform fetch\n", stageImage.Name())

			proc := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from %s", stg.LogDetailedName(), reader.String())
			proc.Start()

			err := doFetchStage(ctx, m.ProjectName, reader, *stageID, stageImage)
			pulled = true

			if storage.IsErrStageNotFound(err) {
				logboek.Context(ctx).Default().LogF("Stage not found\n")
				proc.End()
				return nil, err
			}

			if err != nil {
				proc.Fail()
				return nil, err
			}

			proc.End()

			if err := storeStageDescIntoLocalManifestCache(ctx, m.ProjectName, *stageID, reader, stageImage.GetStageDesc()); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", imageName, err)
			}
		} else {
			logboek.Context(ctx).Info().LogF("Cache repo image %s exists locally, will not perform fetch\n", stageImage.Name())

			stageDesc, err := getStageDesc(ctx, m.ProjectName, *stageID, reader, nil, getStageDescOptions{WithLocalManifestCache: true})
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), reader.String(), err)
			}
			pulled = false
			stageImage.SetStageDesc(stageDesc)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, stageImage.Name()); err != nil {
			return nil, fmt.Errorf("error accessing last recently used images cache for %s: %w", stageImage.Name(), err)
		}

		return stageImage, nil
	}

	prepareCacheStageAsPrimary := func(cacheImg container_backend.LegacyImageInterface, primaryStage stage.Interface) error {
		primaryImg := primaryStage.GetStageImage().Image
		stageID := primaryImg.GetStageDesc().StageID
		primaryImageName := m.primaryReader().ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)

		if err := containerBackend.RenameImage(ctx, cacheImg, primaryImageName, true); err != nil {
			return fmt.Errorf("unable to rename image %s to %s: %w", cacheImg.Name(), primaryImageName, err)
		}

		if err := containerBackend.RefreshImageObject(ctx, primaryImg); err != nil {
			return fmt.Errorf("unable to refresh stage image %s: %w", primaryImageName, err)
		}

		if err := storeStageDescIntoLocalManifestCache(ctx, m.ProjectName, *stageID, m.primaryReader(), primaryImg.GetStageDesc()); err != nil {
			return fmt.Errorf("error storing stage %s description into local manifest cache: %w", primaryImageName, err)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, primaryImageName); err != nil {
			return fmt.Errorf("error accessing last recently used images cache for %s: %w", primaryImageName, err)
		}

		return nil
	}

	for _, cacheReader := range m.CacheReaders {
		cacheImg, err := fetchStageFromCache(cacheReader)
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				unavailableReader = cacheReader
			} else {
				allStagesUnavailable = false
			}

			if !storage.IsErrStageNotFound(err) {
				logboek.Context(ctx).Warn().LogF("Unable to fetch stage %s from cache to storage %s: %s\n", stg.GetStageImage().Image.GetStageDesc().StageID.String(), cacheReader.String(), err)
			}

			if asWriter, ok := cacheReader.(storage.StageWriter); ok {
				writersToRefill = append(writersToRefill, asWriter)
			}

			continue
		}

		allStagesUnavailable = false

		if err := prepareCacheStageAsPrimary(cacheImg, stg); err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to prepare stage %s fetched from cache to storage %s as a primary: %s\n", cacheImg.Name(), cacheReader.String(), err)

			if asWriter, ok := cacheReader.(storage.StageWriter); ok {
				writersToRefill = append(writersToRefill, asWriter)
			}

			continue
		}

		fetchedImg = cacheImg
		fetchedReader = cacheReader
		if sameStageStorage(cacheReader, m.primaryReader()) {
			source = BaseImageSourceTypeRepo
		} else {
			source = BaseImageSourceTypeCacheRepo
		}
		break
	}

	if fetchedImg == nil {
		stageID := stg.GetStageImage().Image.GetStageDesc().StageID

		if allStagesUnavailable {
			if unavailableReader == nil {
				unavailableReader = m.primaryReader()
			}

			logboek.Context(ctx).Error().LogF("Stage %s image %s is no longer available\n", stg.LogDetailedName(), stg.GetStageImage().Image.Name())

			stageImageName := unavailableReader.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)
			if err := image.CommonManifestCache.DeleteImageInfo(ctx, unavailableReader.String(), stageImageName); err != nil {
				logboek.Context(ctx).Warn().LogF("Unable to delete manifest cache for rejected stage %s: %s\n", stageImageName, err)
			}

			if m.MetaStorage != nil {
				logboek.Context(ctx).Error().LogF("Will mark image %s as rejected in the meta storage %s\n", stg.GetStageImage().Image.Name(), m.MetaStorage.String())
				if err := m.MetaStorage.RejectStage(ctx, m.ProjectName, stageID.Digest, stageID.CreationTs); err != nil {
					return FetchStageInfo{}, fmt.Errorf("unable to reject stage %s image %s in the meta storage %s: %w", stg.LogDetailedName(), stg.GetStageImage().Image.Name(), m.MetaStorage.String(), err)
				}
			}

			return FetchStageInfo{}, ErrUnexpectedStorageState
		}

		return FetchStageInfo{}, fmt.Errorf("unable to fetch stage %s from cache readers", stageID.String())
	}

	for _, writer := range writersToRefill {
		stageID := stg.GetStageImage().Image.GetStageDesc().StageID

		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", stg.LogDetailedName(), writer.String()).
			DoError(func() error {
				if _, err := m.CopyStage(ctx, fetchedReader, writer, *stageID, CopyStageOptions{
					ContainerBackend: containerBackend,
					LegacyImage:      fetchedImg,
				}); err != nil {
					return fmt.Errorf("unable to copy stage %s into cache to storage %s: %w", stageID.String(), writer.String(), err)
				}
				return nil
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Warning %s\n", err)
		}
	}

	return FetchStageInfo{BaseImagePulled: pulled, BaseImageSource: source}, nil
}

func (m *StorageManager) CopyStageIntoCacheToStorages(ctx context.Context, stageID image.StageID, cacheToStorages []storage.StageWriter, opts CopyStageIntoStorageOptions) error {
	for _, cache := range cacheToStorages {
		if matchedReader := m.findMatchingReader(cache); matchedReader != nil {
			if stageDesc, err := matchedReader.GetStageDesc(ctx, m.ProjectName, stageID); err == nil && stageDesc != nil {
				logboek.Context(ctx).Debug().LogF("Skip copying stage %s into cache %s: already exists in a cache reader\n", opts.LogDetailedName, cache.String())
				continue
			}
		}

		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", opts.LogDetailedName, cache.String()).
			DoError(func() error {
				copyOpts := CopyStageOptions{ContainerBackend: opts.ContainerBackend}
				if opts.FetchStage != nil {
					copyOpts.FetchStage = opts.FetchStage
					copyOpts.LegacyImage = opts.FetchStage.GetStageImage().Image
				}
				if _, err := m.CopyStage(ctx, m.primaryReader(), cache, stageID, copyOpts); err != nil {
					return fmt.Errorf("unable to copy stage %s into cache to storage %s: %w", stageID.String(), cache.String(), err)
				}
				return nil
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Warning: %s\n", err)
		}
	}
	return nil
}

func (m *StorageManager) findMatchingReader(writer storage.BaseStorage) storage.StageReader {
	for _, reader := range m.CacheReaders {
		if sameStageStorage(reader, writer) {
			return reader
		}
	}
	return nil
}

func (m *StorageManager) getOrCreateImagesRepoStagesListCache(ctx context.Context, imagesRepo storage.ImagesRepoStorage) (*StagesList, error) {
	m.ImagesRepoStagesListCacheMux.Lock()
	defer m.ImagesRepoStagesListCacheMux.Unlock()

	key := imagesRepo.String()
	if m.ImagesRepoStagesListCache == nil {
		m.ImagesRepoStagesListCache = make(map[string]*StagesList)
	}

	if cache, ok := m.ImagesRepoStagesListCache[key]; ok {
		return cache, nil
	}

	stageIDs, err := imagesRepo.GetStagesIDs(ctx, m.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("unable to get images repo stages list from %s: %w", imagesRepo.String(), err)
	}
	cache := NewStagesList(stageIDs)
	m.ImagesRepoStagesListCache[key] = cache

	return cache, nil
}

func (m *StorageManager) CopyStageIntoImagesRepoStorage(ctx context.Context, stageID image.StageID, destImagesRepo storage.ImagesRepoStorage, opts CopyStageIntoStorageOptions) (*image.StageDesc, error) {
	// Check if the stage already exists in the destination images repo
	finalImageName := destImagesRepo.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)

	desc, err := destImagesRepo.GetStageDesc(ctx, m.ProjectName, stageID)
	if err != nil && !storage.IsErrStageUnavailable(err) {
		return nil, fmt.Errorf("unable to get stage %s descriptor from images repo %s: %w", stageID.String(), destImagesRepo.String(), err)
	}
	if desc != nil {
		logboek.Context(ctx).Info().LogF("Stage %s already exists in the images repo %s, skipping\n", stageID.String(), destImagesRepo.String())

		logboek.Context(ctx).Default().LogFHighlight("Use previously built images repo image for %s\n", opts.LogDetailedName)
		container_backend.LogImageName(ctx, finalImageName)

		return desc, nil
	}

	if opts.ShouldBeBuiltMode {
		return nil, fmt.Errorf("%s with digest %s is not exist in the images repo", opts.LogDetailedName, stageID.Digest)
	}

	var stageDescCopy *image.StageDesc
	err = logboek.Context(ctx).Default().LogProcess("Copy stage %s into the images repo", opts.LogDetailedName).
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
			stageDescCopy, err = m.CopyStage(ctx, m.primaryReader(), destImagesRepo, stageID, copyOpts)
			if err != nil {
				return fmt.Errorf("unable to copy stage %s into the images repo %s: %w", stageID.String(), destImagesRepo.String(), err)
			}

			logboek.Context(ctx).Default().LogFDetails("  name: %s\n", finalImageName)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// Update per-repo cache
	if cache, err := m.getOrCreateImagesRepoStagesListCache(ctx, destImagesRepo); err == nil {
		cache.AddStageID(stageID)
	}

	return stageDescCopy, nil
}

func (m *StorageManager) SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, stageDescSet image.StageDescSet) (*image.StageDesc, error) {
	if stageDescSet.IsEmpty() {
		return nil, nil
	}

	var stageDesc *image.StageDesc
	if err := logboek.Context(ctx).Info().LogProcess("Selecting suitable image for stage %s by digest %s", stg.Name(), stg.GetDigest()).
		DoError(func() error {
			var err error
			stageDesc, err = stg.SelectSuitableStageDesc(ctx, c, stageDescSet)
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

func (m *StorageManager) GetStageDescSetByDigestWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error) {
	return m.GetStageDescSetByDigestFromReaderWithCache(ctx, stageName, stageDigest, parentStageCreationTs, m.primaryReader())
}

func (m *StorageManager) GetStageDescSetByDigest(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error) {
	return m.GetStageDescSetByDigestFromReader(ctx, stageName, stageDigest, parentStageCreationTs, m.primaryReader())
}

func (m *StorageManager) GetStageDescSetByDigestFromReaderWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader) (image.StageDescSet, error) {
	cachedStageDescSet, err := m.getStageDescSetByDigestFromReader(ctx, stageName, stageDigest, parentStageCreationTs, reader, storage.WithCache())
	if err != nil {
		return nil, err
	}

	if !cachedStageDescSet.IsEmpty() {
		return cachedStageDescSet, nil
	}

	return m.getStageDescSetByDigestFromReader(ctx, stageName, stageDigest, parentStageCreationTs, reader)
}

func (m *StorageManager) GetStageDescSetByDigestFromReader(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader) (image.StageDescSet, error) {
	return m.getStageDescSetByDigestFromReader(ctx, stageName, stageDigest, parentStageCreationTs, reader)
}

func (m *StorageManager) getStageDescSetByDigestFromReader(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader, opts ...storage.Option) (image.StageDescSet, error) {
	stageIDs, err := m.getStagesIDsByDigestFromReader(ctx, stageName, stageDigest, parentStageCreationTs, reader, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get stages ids from %s by digest %s for stage %s: %w", reader.String(), stageDigest, stageName, err)
	}

	stageDescSet, err := m.getStageDescSetFromReader(ctx, stageIDs, reader, m.otherReaders(reader))
	if err != nil {
		return nil, fmt.Errorf("unable to get stage descriptions by ids from %s: %w", reader.String(), err)
	}

	return stageDescSet, nil
}

func (m *StorageManager) CopySuitableStageDescByDigest(ctx context.Context, stageDesc *image.StageDesc, src storage.StageReader, dest storage.StageWriter, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDesc, error) {
	img := container_backend.NewLegacyStageImage(nil, stageDesc.Info.Name, containerBackend, targetPlatform)

	logboek.Context(ctx).Info().LogF("Fetching %s\n", img.Name())
	if err := src.FetchImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to fetch %s from %s: %w", stageDesc.Info.Name, src.String(), err)
	}

	newImageName := dest.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.CreationTs)
	logboek.Context(ctx).Info().LogF("Renaming image %s to %s\n", img.Name(), newImageName)
	if err := containerBackend.RenameImage(ctx, img, newImageName, false); err != nil {
		return nil, err
	}

	logboek.Context(ctx).Info().LogF("Storing %s\n", newImageName)
	if err := dest.StoreImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to store %s to %s: %w", stageDesc.Info.Name, dest.String(), err)
	}

	destReader, ok := dest.(storage.StageReader)
	if !ok {
		return ConvertStageDescForStorage(stageDesc, dest), nil
	}

	if destStageDesc, err := getStageDesc(ctx, m.ProjectName, *stageDesc.StageID, destReader, m.CacheReaders, getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()}); err != nil {
		return nil, fmt.Errorf("unable to get stage %s description from %s: %w", stageDesc.StageID.String(), dest.String(), err)
	} else {
		return destStageDesc, nil
	}
}

func (m *StorageManager) getWithLocalManifestCacheOption() bool {
	return len(m.CacheReaders) > 0 && m.primaryReader().Address() != storage.LocalStorageAddress
}

func (m *StorageManager) getStagesIDsByDigestFromReader(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, reader storage.StageReader, opts ...storage.Option) ([]image.StageID, error) {
	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Info().LogProcess("Get %s stages by digest %s from storage", stageName, stageDigest).
		DoError(func() error {
			var err error
			stageIDs, err = reader.GetStagesIDsByDigest(ctx, m.ProjectName, stageDigest, parentStageCreationTs, opts...)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from storage: %w", reader.String(), stageDigest, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	return stageIDs, nil
}

func (m *StorageManager) getStageDescSetFromReader(ctx context.Context, stageIDs []image.StageID, reader storage.StageReader, cacheReaders []storage.StageReader) (image.StageDescSet, error) {
	stageDescSet := image.NewStageDescSet()
	for _, stageID := range stageIDs {
		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, reader, cacheReaders, getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), reader.String(), err)
				continue
			}

			return nil, err
		}

		stageDescSet.Add(stageDesc)
	}

	return stageDescSet, nil
}

type getStageDescOptions struct {
	WithLocalManifestCache bool
}

func getStageDescFromLocalManifestCache(ctx context.Context, projectName string, stageID image.StageID, reader storage.StageReader) (*image.StageDesc, error) {
	stageImageName := reader.ConstructStageImageName(projectName, stageID.Digest, stageID.CreationTs)

	logboek.Context(ctx).Debug().LogF("Getting image %s info from the manifest cache...\n", stageImageName)
	imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, reader.String(), stageImageName)
	if err != nil {
		return nil, fmt.Errorf("error getting image %s info: %w", stageImageName, err)
	}

	if imgInfo != nil {
		logboek.Context(ctx).Info().LogF("Got image %s info from the manifest cache (CACHE HIT)\n", stageImageName)

		return &image.StageDesc{
			StageID: image.NewStageID(stageID.Digest, stageID.CreationTs),
			Info:    imgInfo,
		}, nil
	} else {
		logboek.Context(ctx).Info().LogF("Not found %s image info in the manifest cache (CACHE MISS)\n", stageImageName)
	}

	return nil, nil
}

func ConvertStageDescForStorage(stageDesc *image.StageDesc, stg storage.BaseStorage) *image.StageDesc {
	return &image.StageDesc{
		StageID: image.NewStageID(stageDesc.StageID.Digest, stageDesc.StageID.CreationTs),
		Info: &image.Info{
			Name:              fmt.Sprintf("%s:%s-%d", stg.Address(), stageDesc.StageID.Digest, stageDesc.StageID.CreationTs),
			Repository:        stg.Address(),
			Tag:               stageDesc.Info.Tag,
			RepoDigest:        stageDesc.Info.RepoDigest,
			ID:                stageDesc.Info.ID,
			ParentID:          stageDesc.Info.ParentID,
			Labels:            stageDesc.Info.Labels,
			Size:              stageDesc.Info.Size,
			CreatedAtUnixNano: stageDesc.Info.CreatedAtUnixNano,
			OnBuild:           stageDesc.Info.OnBuild,
			Env:               stageDesc.Info.Env,
			Volumes:           stageDesc.Info.Volumes,
		},
	}
}

func getStageDesc(ctx context.Context, projectName string, stageID image.StageID, reader storage.StageReader, cacheReaders []storage.StageReader, opts getStageDescOptions) (*image.StageDesc, error) {
	if opts.WithLocalManifestCache {
		stageDesc, err := getStageDescFromLocalManifestCache(ctx, projectName, stageID, reader)
		if err != nil {
			return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), reader.String(), err)
		}
		if stageDesc != nil {
			return stageDesc, nil
		}
	}

	for _, cacheReader := range cacheReaders {
		if opts.WithLocalManifestCache {
			stageDesc, err := getStageDescFromLocalManifestCache(ctx, projectName, stageID, cacheReader)
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from the local manifest cache: %w", stageID.String(), err)
			}
			if stageDesc != nil {
				return ConvertStageDescForStorage(stageDesc, reader), nil
			}
		}

		var stageDesc *image.StageDesc
		err := logboek.Context(ctx).Info().LogProcess("Get stage %s description from cache to storage %s", stageID.String(), cacheReader.String()).
			DoError(func() error {
				var err error
				stageDesc, err = cacheReader.GetStageDesc(ctx, projectName, stageID)

				logboek.Context(ctx).Debug().LogF("Got stage description: %#v\n", stageDesc)
				return err
			})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				continue
			}

			logboek.Context(ctx).Warn().LogF("Unable to get stage description from cache to storage %s: %s\n", cacheReader.String(), err)
			continue
		}

		if opts.WithLocalManifestCache {
			if err := storeStageDescIntoLocalManifestCache(ctx, projectName, stageID, cacheReader, stageDesc); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
			}
		}

		return ConvertStageDescForStorage(stageDesc, reader), nil
	}

	logboek.Context(ctx).Debug().LogF("Getting digest %q creation timestamp %d stage info from %s...\n", stageID.Digest, stageID.CreationTs, reader.String())
	stageDesc, err := reader.GetStageDesc(ctx, projectName, stageID)
	if err != nil {
		if storage.IsErrStageUnavailable(err) {
			return nil, err
		}

		return nil, fmt.Errorf("error getting digest %q creation timestamp %d stage info from %s: %w", stageID.Digest, stageID.CreationTs, reader.String(), err)
	}

	if opts.WithLocalManifestCache {
		if err := storeStageDescIntoLocalManifestCache(ctx, projectName, stageID, reader, stageDesc); err != nil {
			return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
		}
	}
	return stageDesc, nil
}

func (m *StorageManager) GenerateStageDescCreationTs(digest string, stageDescSet image.StageDescSet) (string, int64) {
	var imageName string

	for {
		timeNow := time.Now().UTC()
		creationTs := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)
		imageName = m.primaryReader().ConstructStageImageName(m.ProjectName, digest, creationTs)

		for stageDesc := range stageDescSet.Iter() {
			if stageDesc.Info.Name == imageName {
				continue
			}
		}

		return imageName, creationTs
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
		if m.MetaStorage == nil {
			return f(ctx, task.commit, task.stageID, nil)
		}
		err := m.MetaStorage.RmImageMetadata(ctx, projectName, imageNameOrID, task.commit, task.stageID)
		return f(ctx, task.commit, task.stageID, err)
	})
}

func (m *StorageManager) ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error {
	return parallel.DoTasks(ctx, len(managedImages), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		managedImage := managedImages[taskId]
		if m.MetaStorage == nil {
			return f(ctx, managedImage, nil)
		}
		err := m.MetaStorage.RmManagedImage(ctx, projectName, managedImage)
		return f(ctx, managedImage, err)
	})
}

func (m *StorageManager) ForEachDeleteStageCustomTag(ctx context.Context, ids []string, f func(ctx context.Context, tag string, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		if len(m.ImagesRepoStorages) == 0 {
			return f(ctx, ids[taskId], fmt.Errorf("images repo storage is not configured"))
		}

		id := ids[taskId]

		// Delete from all images repos
		for _, imgRepo := range m.ImagesRepoStorages {
			if err := imgRepo.DeleteStageCustomTag(ctx, id); err != nil {
				return f(ctx, id, fmt.Errorf("unable to delete stage custom tag: %w", err))
			}
			if err := imgRepo.UnregisterStageCustomTag(ctx, id); err != nil {
				return f(ctx, id, fmt.Errorf("unable to unregister stage custom tag: %w", err))
			}
		}

		return f(ctx, id, nil)
	})
}

func (m *StorageManager) ForEachGetStageCustomTagMetadata(ctx context.Context, ids []string, f func(ctx context.Context, metadataID string, metadata *storage.CustomTagMetadata, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		if len(m.ImagesRepoStorages) == 0 {
			return f(ctx, ids[taskId], nil, fmt.Errorf("images repo storage is not configured"))
		}

		id := ids[taskId]

		// Try to get metadata from the first images repo that has it
		var lastErr error
		for _, imgRepo := range m.ImagesRepoStorages {
			metadata, err := imgRepo.GetStageCustomTagMetadata(ctx, id)
			if err == nil {
				return f(ctx, id, metadata, nil)
			}
			if !storage.IsErrCustomTagMetadataNotFound(err) {
				lastErr = err
			}
		}

		if lastErr != nil {
			return f(ctx, id, nil, lastErr)
		}

		return f(ctx, id, nil, nil)
	})
}
