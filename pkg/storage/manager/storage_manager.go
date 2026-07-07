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
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var ErrUnexpectedStagesStorageState = errors.New("unexpected stages storage state")

const maxRetryAttemptsOnUnexpectedStagesStorageState = 4

func IsErrUnexpectedStagesStorageState(err error) bool {
	if err != nil {
		return strings.HasSuffix(err.Error(), ErrUnexpectedStagesStorageState.Error())
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

	GetStagesStorage() storage.RegistryStorage
	GetMetaStorage() storage.RegistryStorage
	GetFinalImagesStorage() storage.RegistryStorage
	GetImagesStorage() storage.RegistryStorage
	GetCustomTagsStorage() storage.RegistryStorage
	IsRemoteImagesStorage() bool
	GetSecondaryStagesStorageList() []storage.RegistryStorage
	GetCacheStagesStorageList() []storage.RegistryStorage
	GetCacheStagesWriteList() []storage.RegistryStorage

	GetImageInfoGetter(imageName string, desc *image.StageDesc, opts image.InfoGetterOptions) *image.InfoGetter

	EnableParallel(parallelTasksLimit int)
	MaxNumberOfWorkers() int
	GenerateStageDescCreationTs(digest string, stageDescSet image.StageDescSet) (string, int64)

	LockStageImage(ctx context.Context, imageName string) error
	GetStageDescSetByDigest(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error)
	GetStageDescSetByDigestWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error)
	GetStageDescSetByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage) (image.StageDescSet, error)
	GetStageDescSetByDigestFromStagesStorageWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage) (image.StageDescSet, error)
	GetStageDescSet(ctx context.Context) (image.StageDescSet, error)
	GetStageDescSetWithCache(ctx context.Context) (image.StageDescSet, error)
	GetFinalStageDescSet(ctx context.Context) (image.StageDescSet, error)

	FetchStage(ctx context.Context, containerBackend container_backend.ContainerBackend, stg stage.Interface) (FetchStageInfo, error)
	FetchStageImage(ctx context.Context, containerBackend container_backend.ContainerBackend, logName string, stageImage *stage.StageImage) (FetchStageInfo, error)
	SelectSuitableStageDesc(ctx context.Context, c stage.Conveyor, stg stage.Interface, stageDescSet image.StageDescSet) (*image.StageDesc, error)
	CopySuitableStageDescByDigest(ctx context.Context, stageDesc *image.StageDesc, sourceRegistryStorage, destinationRegistryStorage storage.RegistryStorage, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDesc, error)
	CopyStageIntoCacheStorages(ctx context.Context, stageID image.StageID, cacheStagesStorages []storage.RegistryStorage, opts CopyStageIntoStorageOptions) error
	CopyStageIntoFinalStorage(ctx context.Context, stageID image.StageID, finalImagesStorage storage.RegistryStorage, opts CopyStageIntoStorageOptions) (*image.StageDesc, error)

	ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error
	ForEachDeleteFinalStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error
	ForEachRejectedStage(ctx context.Context, stageIDs []image.StageID, f func(ctx context.Context, stageID image.StageID) error) error
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

func RetryOnUnexpectedStagesStorageState(ctx context.Context, _ StorageManagerInterface, f func() error) error {
	var attempt int
	op := func() (bool, error) {
		attempt++

		if err := f(); err != nil {
			if !IsErrUnexpectedStagesStorageState(err) {
				return false, backoff.Permanent(err)
			}
			return false, fmt.Errorf("exhausted %d retries on unexpected stages storage state: %w", maxRetryAttemptsOnUnexpectedStagesStorageState-1, err)
		}

		return false, nil
	}

	notify := func(err error, duration time.Duration) {
		logboek.Context(ctx).Warn().LogF("Retrying due to unexpected stages storage state (attempt %d/%d) in %0.2f seconds ...\n", attempt, maxRetryAttemptsOnUnexpectedStagesStorageState-1, duration.Seconds())
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = 2 * time.Second
	eb.MaxInterval = 10 * time.Second

	_, err := backoff.Retry(ctx, op,
		backoff.WithBackOff(eb),
		backoff.WithMaxTries(maxRetryAttemptsOnUnexpectedStagesStorageState),
		backoff.WithNotify(notify),
	)
	return err
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

	// Storages groups every repo/registry in use under the granular registry
	// model (--repo preset or --cache-from/--cache-to/--images-repo/
	// --meta-repo/--final-repo). See Storages for field-by-field docs.
	Storages Storages

	// These will be released automatically when current process exits
	SharedHostImagesLocks []lockgate.LockHandle

	FinalImageListCacheMux sync.Mutex
	FinalImageListCache    *StagesList
}

func (m *StorageManager) GetStagesStorage() storage.RegistryStorage {
	return m.Storages.Stages
}

// LogRepositoriesUsed prints the effective repositories in use at the start of
// a run, so the resolved registry model (from --repo preset, aliases or the
// granular flags) is visible for debugging.
func (m *StorageManager) LogRepositoriesUsed(ctx context.Context) {
	addrs := func(list []storage.RegistryStorage) string {
		if len(list) == 0 {
			return "-"
		}
		res := make([]string, 0, len(list))
		for _, s := range list {
			res = append(res, s.Address())
		}
		return strings.Join(res, ", ")
	}

	imagesRepo := "-"
	if m.Storages.Images != nil {
		imagesRepo = m.Storages.Images.Address()
	}

	// cache-from and cache-to have no explicit read/write list under the
	// --repo preset — stages are read from and written to the primary storage
	// directly — so display the effective primary address rather than "-"
	// when the corresponding list is empty.
	cacheFrom := addrs(m.Storages.CacheFrom)
	if len(m.Storages.CacheFrom) == 0 {
		cacheFrom = m.Storages.Stages.Address()
	}
	cacheTo := addrs(m.Storages.CacheTo)
	if len(m.Storages.CacheTo) == 0 {
		cacheTo = m.Storages.Stages.Address()
	}

	logboek.Context(ctx).Default().LogBlock("Repositories").Do(func() {
		logboek.Context(ctx).Default().LogF("stages:      %s\n", m.Storages.Stages.Address())
		logboek.Context(ctx).Default().LogF("images-repo: %s\n", imagesRepo)
		if m.Storages.Final != nil {
			logboek.Context(ctx).Default().LogF("final-repo:  %s\n", m.Storages.Final.Address())
		}
		logboek.Context(ctx).Default().LogF("cache-from:  %s\n", cacheFrom)
		logboek.Context(ctx).Default().LogF("cache-to:    %s\n", cacheTo)
		logboek.Context(ctx).Default().LogF("meta-repo:   %s\n", m.GetMetaStorage().Address())
	})
}

func (m *StorageManager) GetFinalImagesStorage() storage.RegistryStorage {
	return m.Storages.Final
}

func (m *StorageManager) GetImagesStorage() storage.RegistryStorage {
	return m.Storages.Images
}

// GetCustomTagsStorage returns the storage holding custom-tag alias images:
// the final images storage when set, otherwise the content-tag images
// storage. Matches the publish path so deletion targets the same repo.
func (m *StorageManager) GetCustomTagsStorage() storage.RegistryStorage {
	return m.Storages.CustomTags()
}

// IsRemoteImagesStorage reports whether the content-tag storage
// (--images-repo) is a remote registry, as opposed to :local. Image metadata
// publishing (managed-image records, custom tags, git metadata) only makes
// sense for a remote registry, so callers use this as the single guard
// deciding whether to run metadata publishing at all.
func (m *StorageManager) IsRemoteImagesStorage() bool {
	return m.Storages.IsRemoteImagesStorage()
}

func (m *StorageManager) GetSecondaryStagesStorageList() []storage.RegistryStorage {
	return m.Storages.Secondary
}

func (m *StorageManager) GetCacheStagesStorageList() []storage.RegistryStorage {
	return m.Storages.CacheFrom
}

func (m *StorageManager) GetCacheStagesWriteList() []storage.RegistryStorage {
	return m.Storages.CacheTo
}

// GetMetaStorage returns the storage that holds build/cleanup metadata.
// Falls back to the primary stages storage when no dedicated meta repo is set
// (i.e. the --repo preset), preserving co-located behavior bit-for-bit.
func (m *StorageManager) GetMetaStorage() storage.RegistryStorage {
	return m.Storages.Meta()
}

func (m *StorageManager) GetServiceValuesRepo() string {
	if m.Storages.Final != nil {
		return m.Storages.Final.String()
	}
	return m.Storages.Stages.String()
}

func (m *StorageManager) GetImageInfoGetter(imageName string, stageDesc *image.StageDesc, opts image.InfoGetterOptions) *image.InfoGetter {
	if m.Storages.Final != nil {
		finalImageName := m.Storages.Final.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.CreationTs)
		return image.NewInfoGetter(imageName, finalImageName, stageDesc.Info.GetDigest(), opts)
	}

	return image.NewInfoGetter(imageName, stageDesc.Info.Name, stageDesc.Info.GetDigest(), opts)
}

func (m *StorageManager) InitCache(ctx context.Context) error {
	logboek.Context(ctx).Info().LogF("Initializing storage manager cache\n")

	if m.Storages.Final != nil {
		if _, err := m.getOrCreateFinalImageListCache(ctx); err != nil {
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

func (m *StorageManager) GetStageDescSetWithCache(ctx context.Context) (image.StageDescSet, error) {
	return m.getStageDescSet(ctx, storage.WithCache())
}

func (m *StorageManager) GetStageDescSet(ctx context.Context) (image.StageDescSet, error) {
	return m.getStageDescSet(ctx)
}

func (m *StorageManager) getStageDescSet(ctx context.Context, opts ...storage.Option) (image.StageDescSet, error) {
	stageIDs, err := m.Storages.Stages.GetStagesIDs(ctx, m.ProjectName, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting stages ids from %s: %w", m.Storages.Stages, err)
	}

	stageDescSet := image.NewStageDescSet()
	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, m.Storages.Stages, m.Storages.CacheFrom, getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), m.Storages.Stages.String(), err)
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

func (m *StorageManager) GetFinalStageDescSet(ctx context.Context) (image.StageDescSet, error) {
	existingStagesListCache, err := m.getOrCreateFinalImageListCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting existing stages list of final repo %s: %w", m.Storages.Final.String(), err)
	}

	logboek.Context(ctx).Debug().LogF("[%p] Got existing final stages list cache: %#v\n", m, existingStagesListCache.StageIDs)

	stageIDs := existingStagesListCache.GetStageIDs()
	stageDescSet := image.NewStageDescSet()
	if err := parallel.DoTasks(ctx, len(stageIDs), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		stageID := stageIDs[taskId]

		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, m.Storages.Final, nil, getStageDescOptions{WithLocalManifestCache: true})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), m.Storages.Final.String(), err)
				return nil
			}

			return fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), m.Storages.Final.String(), err)
		}

		stageDescSet.Add(stageDesc)

		return nil
	}); err != nil {
		return nil, err
	}

	return stageDescSet, nil
}

func (m *StorageManager) ForEachDeleteFinalStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error {
	stageDescSet = stageDescSet.Clone()
	return parallel.DoTasks(ctx, stageDescSet.Cardinality(), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDesc, _ := stageDescSet.Pop()
		err := m.Storages.Final.DeleteStage(ctx, stageDesc, options.DeleteImageOptions)
		return f(ctx, stageDesc, err)
	})
}

func (m *StorageManager) ForEachDeleteStage(ctx context.Context, options ForEachDeleteStageOptions, stageDescSet image.StageDescSet, f func(ctx context.Context, stageDesc *image.StageDesc, err error) error) error {
	if localRegistryStorage, isLocal := m.Storages.Stages.(*storage.LocalRegistryStorage); isLocal {
		filteredStageDescSet, err := localRegistryStorage.FilterStageDescSetAndProcessRelatedData(ctx, stageDescSet, options.FilterStagesAndProcessRelatedDataOptions)
		if err != nil {
			return fmt.Errorf("error filtering local docker server stages: %w", err)
		}

		stageDescSet = filteredStageDescSet
	}

	stageDescSet = stageDescSet.Clone()
	return parallel.DoTasks(ctx, stageDescSet.Cardinality(), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		stageDesc, _ := stageDescSet.Pop()

		for _, cacheRegistryStorage := range m.Storages.CacheFrom {
			if err := cacheRegistryStorage.DeleteStage(ctx, stageDesc, options.DeleteImageOptions); err != nil {
				logboek.Context(ctx).Warn().LogF("Unable to delete stage %s from the cache stages storage %s: %s\n", stageDesc.StageID.String(), cacheRegistryStorage.String(), err)
			}
		}

		err := m.Storages.Stages.DeleteStage(ctx, stageDesc, options.DeleteImageOptions)
		return f(ctx, stageDesc, err)
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

func doFetchStage(ctx context.Context, projectName string, registryStorage storage.RegistryStorage, stageID image.StageID, img container_backend.LegacyImageInterface) error {
	err := logboek.Context(ctx).Info().LogProcess("Check manifest availability").DoError(func() error {
		freshStageDesc, err := registryStorage.GetStageDesc(ctx, projectName, stageID)
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

		if err := registryStorage.FetchImage(ctx, img); err != nil {
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
	return m.FetchStageImage(ctx, containerBackend, stg.LogDetailedName(), stg.GetStageImage())
}

func (m *StorageManager) FetchStageImage(ctx context.Context, containerBackend container_backend.ContainerBackend, logName string, stageImage *stage.StageImage) (FetchStageInfo, error) {
	logboek.Context(ctx).Debug().LogF("-- StagesManager.FetchStage %s\n", logName)

	if err := m.LockStageImage(ctx, stageImage.Image.Name()); err != nil {
		return FetchStageInfo{}, fmt.Errorf("error locking stage image %q: %w", stageImage.Image.Name(), err)
	}

	shouldFetch, err := m.Storages.Stages.ShouldFetchImage(ctx, stageImage.Image)
	if err != nil {
		return FetchStageInfo{}, fmt.Errorf("error checking should fetch image: %w", err)
	}
	if !shouldFetch {
		imageName := m.Storages.Stages.ConstructStageImageName(m.ProjectName, stageImage.Image.GetStageDesc().StageID.Digest, stageImage.Image.GetStageDesc().StageID.CreationTs)

		logboek.Context(ctx).Info().LogF("Image %s exists, will not perform fetch\n", imageName)

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, imageName); err != nil {
			return FetchStageInfo{}, fmt.Errorf("error accessing last recently used images cache for %s: %w", imageName, err)
		}

		return FetchStageInfo{BaseImagePulled: false}, nil
	}

	var fetchedImg container_backend.LegacyImageInterface
	var cacheStagesStorageListToRefill []storage.RegistryStorage
	var pulled bool
	var source string

	fetchStageFromCache := func(registryStorage storage.RegistryStorage) (container_backend.LegacyImageInterface, error) {
		stageID := stageImage.Image.GetStageDesc().StageID
		imageName := registryStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)
		cacheStageImage := container_backend.NewLegacyStageImage(nil, imageName, containerBackend, stageImage.Image.GetTargetPlatform())

		shouldFetch, err := registryStorage.ShouldFetchImage(ctx, cacheStageImage)
		if err != nil {
			return nil, fmt.Errorf("error checking should fetch image from cache repo %s: %w", registryStorage.String(), err)
		}

		if shouldFetch {
			logboek.Context(ctx).Info().LogF("Cache repo image %s does not exist locally, will perform fetch\n", cacheStageImage.Name())

			proc := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from %s", logName, registryStorage.String())
			proc.Start()

			err := doFetchStage(ctx, m.ProjectName, registryStorage, *stageID, cacheStageImage)
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

			if err := storeStageDescIntoLocalManifestCache(ctx, m.ProjectName, *stageID, registryStorage, cacheStageImage.GetStageDesc()); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", imageName, err)
			}
		} else {
			logboek.Context(ctx).Info().LogF("Cache repo image %s exists locally, will not perform fetch\n", cacheStageImage.Name())

			stageDesc, err := getStageDesc(ctx, m.ProjectName, *stageID, registryStorage, nil, getStageDescOptions{WithLocalManifestCache: true})
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), m.Storages.Final.String(), err)
			}
			pulled = false
			cacheStageImage.SetStageDesc(stageDesc)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, cacheStageImage.Name()); err != nil {
			return nil, fmt.Errorf("error accessing last recently used images cache for %s: %w", cacheStageImage.Name(), err)
		}

		return cacheStageImage, nil
	}

	prepareCacheStageAsPrimary := func(cacheImg, primaryImg container_backend.LegacyImageInterface) error {
		stageID := primaryImg.GetStageDesc().StageID
		primaryImageName := m.Storages.Stages.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)

		if err := containerBackend.RenameImage(ctx, cacheImg, primaryImageName, true); err != nil {
			return fmt.Errorf("unable to rename image %s to %s: %w", cacheImg.Name(), primaryImageName, err)
		}

		if err := containerBackend.RefreshImageObject(ctx, primaryImg); err != nil {
			return fmt.Errorf("unable to refresh stage image %s: %w", primaryImageName, err)
		}

		if err := storeStageDescIntoLocalManifestCache(ctx, m.ProjectName, *stageID, m.Storages.Stages, primaryImg.GetStageDesc()); err != nil {
			return fmt.Errorf("error storing stage %s description into local manifest cache: %w", primaryImageName, err)
		}

		if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, primaryImageName); err != nil {
			return fmt.Errorf("error accessing last recently used images cache for %s: %w", primaryImageName, err)
		}

		return nil
	}

	for _, cacheRegistryStorage := range m.Storages.CacheFrom {
		cacheImg, err := fetchStageFromCache(cacheRegistryStorage)
		if err != nil {
			if !storage.IsErrStageNotFound(err) {
				logboek.Context(ctx).Warn().LogF("Unable to fetch stage %s from cache stages storage %s: %s\n", stageImage.Image.GetStageDesc().StageID.String(), cacheRegistryStorage.String(), err)
			}

			cacheStagesStorageListToRefill = append(cacheStagesStorageListToRefill, cacheRegistryStorage)

			continue
		}

		if err := prepareCacheStageAsPrimary(cacheImg, stageImage.Image); err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to prepare stage %s fetched from cache stages storage %s as a primary: %s\n", cacheImg.Name(), cacheRegistryStorage.String(), err)

			cacheStagesStorageListToRefill = append(cacheStagesStorageListToRefill, cacheRegistryStorage)

			continue
		}

		fetchedImg = cacheImg
		source = BaseImageSourceTypeCacheRepo
		break
	}

	if fetchedImg == nil {
		stageID := stageImage.Image.GetStageDesc().StageID
		img := stageImage

		err := logboek.Context(ctx).Default().LogProcess("Fetching stage %s from %s", logName, m.Storages.Stages.String()).
			DoError(func() error {
				return doFetchStage(ctx, m.ProjectName, m.Storages.Stages, *stageID, img.Image)
			})

		if storage.IsErrStageUnavailable(err) {
			logboek.Context(ctx).Error().LogF("Stage %s image %s is no longer available: %s!\n", logName, stageImage.Image.Name(), err)

			// Invalidate manifest cache for the rejected stage (do this regardless of RejectStage result)
			stageImageName := m.Storages.Stages.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)
			if err := image.CommonManifestCache.DeleteImageInfo(ctx, m.Storages.Stages.String(), stageImageName); err != nil {
				logboek.Context(ctx).Warn().LogF("Unable to delete manifest cache for rejected stage %s: %s\n", stageImageName, err)
			}

			logboek.Context(ctx).Error().LogF("Will mark image %s as rejected in the stages storage %s\n", stageImage.Image.Name(), m.Storages.Stages.String())
			if err := m.Storages.Stages.RejectStage(ctx, m.ProjectName, stageID.Digest, stageID.CreationTs); err != nil {
				return FetchStageInfo{}, fmt.Errorf("unable to reject stage %s image %s in the stages storage %s: %w", logName, stageImage.Image.Name(), m.Storages.Stages.String(), err)
			}

			return FetchStageInfo{}, ErrUnexpectedStagesStorageState
		}

		if err != nil {
			return FetchStageInfo{}, fmt.Errorf("unable to fetch stage %s from stages storage %s: %w", stageID.String(), m.Storages.Stages.String(), err)
		}

		source = BaseImageSourceTypeRepo
		fetchedImg = img.Image
	}

	for _, cacheRegistryStorage := range cacheStagesStorageListToRefill {
		stageID := stageImage.Image.GetStageDesc().StageID

		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", logName, cacheRegistryStorage.String()).
			DoError(func() error {
				if _, err := m.CopyStage(ctx, m.Storages.Stages, cacheRegistryStorage, *stageID, CopyStageOptions{
					ContainerBackend: containerBackend,
					LegacyImage:      fetchedImg,
				}); err != nil {
					return fmt.Errorf("unable to copy stage %s into cache stages storage %s: %w", stageID.String(), cacheRegistryStorage.String(), err)
				}
				return nil
			})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Warning %s\n", err)
		}
	}

	return FetchStageInfo{BaseImagePulled: pulled, BaseImageSource: source}, nil
}

func (m *StorageManager) CopyStageIntoCacheStorages(ctx context.Context, stageID image.StageID, cacheStagesStorageList []storage.RegistryStorage, opts CopyStageIntoStorageOptions) error {
	for _, cache := range cacheStagesStorageList {
		err := logboek.Context(ctx).Default().LogProcess("Copy stage %s into cache %s", opts.LogDetailedName, cache.String()).
			DoError(func() error {
				copyOpts := CopyStageOptions{ContainerBackend: opts.ContainerBackend}
				if opts.FetchStage != nil {
					copyOpts.FetchStage = opts.FetchStage
					copyOpts.LegacyImage = opts.FetchStage.GetStageImage().Image
				}
				if _, err := m.CopyStage(ctx, m.Storages.Stages, cache, stageID, copyOpts); err != nil {
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

func (m *StorageManager) getOrCreateFinalImageListCache(ctx context.Context) (*StagesList, error) {
	m.FinalImageListCacheMux.Lock()
	defer m.FinalImageListCacheMux.Unlock()

	if m.FinalImageListCache != nil {
		return m.FinalImageListCache, nil
	}

	stageIDs, err := m.Storages.Final.GetStagesIDs(ctx, m.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("unable to get final repo stages list: %w", err)
	}
	m.FinalImageListCache = NewStagesList(stageIDs)

	return m.FinalImageListCache, nil
}

func (m *StorageManager) CopyStageIntoFinalStorage(ctx context.Context, stageID image.StageID, finalImagesStorage storage.RegistryStorage, opts CopyStageIntoStorageOptions) (*image.StageDesc, error) {
	existingStagesListCache, err := m.getOrCreateFinalImageListCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting existing stages list of final repo %s: %w", finalImagesStorage.String(), err)
	}

	logboek.Context(ctx).Debug().LogF("[%p] Got existing final stages list cache: %#v\n", m, existingStagesListCache.StageIDs)

	finalImageName := finalImagesStorage.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)

	for _, existingStg := range existingStagesListCache.GetStageIDs() {
		if existingStg.IsEqual(stageID) {
			desc, err := m.GetFinalImagesStorage().GetStageDesc(ctx, m.ProjectName, stageID)
			if err != nil && !storage.IsErrStageUnavailable(err) {
				return nil, fmt.Errorf("unable to get stage %s descriptor from final repo %s: %w", stageID.String(), m.GetFinalImagesStorage().String(), err)
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

	var stageDescCopy *image.StageDesc
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
			stageDescCopy, err = m.CopyStage(ctx, m.Storages.Images, finalImagesStorage, stageID, copyOpts)
			if err != nil {
				return fmt.Errorf("unable to copy stage %s into the final repo %s: %w", stageID.String(), finalImagesStorage.String(), err)
			}

			logboek.Context(ctx).Default().LogFDetails("  name: %s\n", finalImageName)

			return nil
		})
	if err != nil {
		return nil, err
	}

	existingStagesListCache.AddStageID(stageID)
	logboek.Context(ctx).Debug().LogF("Updated existing final stages list: %#v\n", m.FinalImageListCache.StageIDs)

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
	return m.GetStageDescSetByDigestFromStagesStorageWithCache(ctx, stageName, stageDigest, parentStageCreationTs, m.Storages.Stages)
}

func (m *StorageManager) GetStageDescSetByDigest(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64) (image.StageDescSet, error) {
	return m.GetStageDescSetByDigestFromStagesStorage(ctx, stageName, stageDigest, parentStageCreationTs, m.Storages.Stages)
}

func (m *StorageManager) GetStageDescSetByDigestFromStagesStorageWithCache(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage) (image.StageDescSet, error) {
	cachedStageDescSet, err := m.getStageDescSetByDigestFromStagesStorage(ctx, stageName, stageDigest, parentStageCreationTs, registryStorage, storage.WithCache())
	if err != nil {
		return nil, err
	}

	if !cachedStageDescSet.IsEmpty() {
		return cachedStageDescSet, nil
	}

	return m.getStageDescSetByDigestFromStagesStorage(ctx, stageName, stageDigest, parentStageCreationTs, registryStorage)
}

func (m *StorageManager) GetStageDescSetByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage) (image.StageDescSet, error) {
	return m.getStageDescSetByDigestFromStagesStorage(ctx, stageName, stageDigest, parentStageCreationTs, registryStorage)
}

func (m *StorageManager) getStageDescSetByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage, opts ...storage.Option) (image.StageDescSet, error) {
	stageIDs, err := m.getStagesIDsByDigestFromStagesStorage(ctx, stageName, stageDigest, parentStageCreationTs, registryStorage, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to get stages ids from %s by digest %s for stage %s: %w", registryStorage.String(), stageDigest, stageName, err)
	}

	stageDescSet, err := m.getStageDescSetFromStagesStorage(ctx, stageIDs, registryStorage, m.Storages.CacheFrom)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage descriptions by ids from %s: %w", registryStorage.String(), err)
	}

	return stageDescSet, nil
}

func (m *StorageManager) CopySuitableStageDescByDigest(ctx context.Context, stageDesc *image.StageDesc, sourceRegistryStorage, destinationRegistryStorage storage.RegistryStorage, containerBackend container_backend.ContainerBackend, targetPlatform string) (*image.StageDesc, error) {
	img := container_backend.NewLegacyStageImage(nil, stageDesc.Info.Name, containerBackend, targetPlatform)

	logboek.Context(ctx).Info().LogF("Fetching %s\n", img.Name())
	if err := sourceRegistryStorage.FetchImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to fetch %s from %s: %w", stageDesc.Info.Name, sourceRegistryStorage.String(), err)
	}

	newImageName := destinationRegistryStorage.ConstructStageImageName(m.ProjectName, stageDesc.StageID.Digest, stageDesc.StageID.CreationTs)
	logboek.Context(ctx).Info().LogF("Renaming image %s to %s\n", img.Name(), newImageName)
	if err := containerBackend.RenameImage(ctx, img, newImageName, false); err != nil {
		return nil, err
	}

	logboek.Context(ctx).Info().LogF("Storing %s\n", newImageName)
	if err := destinationRegistryStorage.StoreImage(ctx, img); err != nil {
		return nil, fmt.Errorf("unable to store %s to %s: %w", stageDesc.Info.Name, destinationRegistryStorage.String(), err)
	}

	if destinationStageDesc, err := getStageDesc(ctx, m.ProjectName, *stageDesc.StageID, destinationRegistryStorage, m.Storages.CacheFrom, getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()}); err != nil {
		return nil, fmt.Errorf("unable to get stage %s description from %s: %w", stageDesc.StageID.String(), destinationRegistryStorage.String(), err)
	} else {
		return destinationStageDesc, nil
	}
}

func (m *StorageManager) getWithLocalManifestCacheOption() bool {
	return m.Storages.Stages.Address() != storage.LocalStorageAddress
}

func (m *StorageManager) getStagesIDsByDigestFromStagesStorage(ctx context.Context, stageName, stageDigest string, parentStageCreationTs int64, registryStorage storage.RegistryStorage, opts ...storage.Option) ([]image.StageID, error) {
	var stageIDs []image.StageID
	if err := logboek.Context(ctx).Info().LogProcess("Get %s stages by digest %s from storage", stageName, stageDigest).
		DoError(func() error {
			var err error
			stageIDs, err = registryStorage.GetStagesIDsByDigest(ctx, m.ProjectName, stageDigest, parentStageCreationTs, opts...)
			if err != nil {
				return fmt.Errorf("error getting project %s stage %s images from storage: %w", m.Storages.Stages.String(), stageDigest, err)
			}

			logboek.Context(ctx).Debug().LogF("Stages ids: %#v\n", stageIDs)

			return nil
		}); err != nil {
		return nil, err
	}

	return stageIDs, nil
}

func (m *StorageManager) getStageDescSetFromStagesStorage(ctx context.Context, stageIDs []image.StageID, registryStorage storage.RegistryStorage, cacheStagesStorageList []storage.RegistryStorage) (image.StageDescSet, error) {
	stageDescSet := image.NewStageDescSet()
	for _, stageID := range stageIDs {
		stageDesc, err := getStageDesc(ctx, m.ProjectName, stageID, registryStorage, cacheStagesStorageList, getStageDescOptions{WithLocalManifestCache: m.getWithLocalManifestCacheOption()})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				logboek.Context(ctx).Warn().LogF("Ignoring stage %s: cannot get stage description from %s: %s\n", stageID.String(), m.Storages.Stages.String(), err)
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

func getStageDescFromLocalManifestCache(ctx context.Context, projectName string, stageID image.StageID, registryStorage storage.RegistryStorage) (*image.StageDesc, error) {
	stageImageName := registryStorage.ConstructStageImageName(projectName, stageID.Digest, stageID.CreationTs)

	logboek.Context(ctx).Debug().LogF("Getting image %s info from the manifest cache...\n", stageImageName)
	imgInfo, err := image.CommonManifestCache.GetImageInfo(ctx, registryStorage.String(), stageImageName)
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

func ConvertStageDescForStagesStorage(projectName string, stageDesc *image.StageDesc, registryStorage storage.RegistryStorage) *image.StageDesc {
	return &image.StageDesc{
		StageID: image.NewStageID(stageDesc.StageID.Digest, stageDesc.StageID.CreationTs),
		Info: &image.Info{
			Name:              registryStorage.ConstructStageImageName(projectName, stageDesc.StageID.Digest, stageDesc.StageID.CreationTs),
			Repository:        registryStorage.Address(),
			Tag:               stageDesc.Info.Tag,
			RepoDigest:        stageDesc.Info.RepoDigest,
			ID:                stageDesc.Info.ID,
			Labels:            stageDesc.Info.Labels,
			Size:              stageDesc.Info.Size,
			CreatedAtUnixNano: stageDesc.Info.CreatedAtUnixNano,
			OnBuild:           stageDesc.Info.OnBuild,
			Env:               stageDesc.Info.Env,
			Volumes:           stageDesc.Info.Volumes,
		},
	}
}

func getStageDesc(ctx context.Context, projectName string, stageID image.StageID, registryStorage storage.RegistryStorage, cacheStagesStorageList []storage.RegistryStorage, opts getStageDescOptions) (*image.StageDesc, error) {
	if opts.WithLocalManifestCache {
		stageDesc, err := getStageDescFromLocalManifestCache(ctx, projectName, stageID, registryStorage)
		if err != nil {
			return nil, fmt.Errorf("error getting stage %s description from %s: %w", stageID.String(), registryStorage.String(), err)
		}
		if stageDesc != nil {
			return stageDesc, nil
		}
	}

	for _, cacheRegistryStorage := range cacheStagesStorageList {
		if opts.WithLocalManifestCache {
			stageDesc, err := getStageDescFromLocalManifestCache(ctx, projectName, stageID, cacheRegistryStorage)
			if err != nil {
				return nil, fmt.Errorf("error getting stage %s description from the local manifest cache: %w", stageID.String(), err)
			}
			if stageDesc != nil {
				return ConvertStageDescForStagesStorage(projectName, stageDesc, registryStorage), nil
			}
		}

		var stageDesc *image.StageDesc
		err := logboek.Context(ctx).Info().LogProcess("Get stage %s description from cache stages storage %s", stageID.String(), cacheRegistryStorage.String()).
			DoError(func() error {
				var err error
				stageDesc, err = cacheRegistryStorage.GetStageDesc(ctx, projectName, stageID)

				logboek.Context(ctx).Debug().LogF("Got stage description: %#v\n", stageDesc)
				return err
			})
		if err != nil {
			if storage.IsErrStageUnavailable(err) {
				continue
			}

			logboek.Context(ctx).Warn().LogF("Unable to get stage description from cache stages storage %s: %s\n", cacheRegistryStorage.String(), err)
			continue
		}

		if stageDesc == nil {
			continue
		}

		if opts.WithLocalManifestCache {
			if err := storeStageDescIntoLocalManifestCache(ctx, projectName, stageID, cacheRegistryStorage, stageDesc); err != nil {
				return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
			}
		}

		return ConvertStageDescForStagesStorage(projectName, stageDesc, registryStorage), nil
	}

	logboek.Context(ctx).Debug().LogF("Getting digest %q creation timestamp %d stage info from %s...\n", stageID.Digest, stageID.CreationTs, registryStorage.String())
	stageDesc, err := registryStorage.GetStageDesc(ctx, projectName, stageID)
	if err != nil {
		if storage.IsErrStageUnavailable(err) {
			return nil, err
		}

		return nil, fmt.Errorf("error getting digest %q creation timestamp %d stage info from %s: %w", stageID.Digest, stageID.CreationTs, registryStorage.String(), err)
	}

	if opts.WithLocalManifestCache {
		if err := storeStageDescIntoLocalManifestCache(ctx, projectName, stageID, registryStorage, stageDesc); err != nil {
			return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", stageID.String(), err)
		}
	}
	return stageDesc, nil
}

func (m *StorageManager) GenerateStageDescCreationTs(digest string, stageDescSet image.StageDescSet) (string, int64) {
	return GenerateStageDescCreationTsForStorage(m.Storages.Stages, m.ProjectName, digest, stageDescSet)
}

// GenerateStageDescCreationTsForStorage picks a creation timestamp whose
// resulting image name does not collide with any stage already in the set.
func GenerateStageDescCreationTsForStorage(registryStorage storage.RegistryStorage, projectName, digest string, stageDescSet image.StageDescSet) (string, int64) {
	timeNow := time.Now().UTC()
	creationTs := timeNow.Unix()*1000 + int64(timeNow.Nanosecond()/1000000)

	for {
		imageName := registryStorage.ConstructStageImageName(projectName, digest, creationTs)

		collision := false
		for stageDesc := range stageDescSet.Iter() {
			if stageDesc.Info.Name == imageName {
				collision = true
				break
			}
		}

		if !collision {
			return imageName, creationTs
		}

		creationTs++
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
		err := m.GetMetaStorage().RmImageMetadata(ctx, projectName, imageNameOrID, task.commit, task.stageID)
		return f(ctx, task.commit, task.stageID, err)
	})
}

func (m *StorageManager) ForEachRmManagedImage(ctx context.Context, projectName string, managedImages []string, f func(ctx context.Context, managedImage string, err error) error) error {
	return parallel.DoTasks(ctx, len(managedImages), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		managedImage := managedImages[taskId]
		err := m.GetMetaStorage().RmManagedImage(ctx, projectName, managedImage)
		return f(ctx, managedImage, err)
	})
}

func (m *StorageManager) ForEachRejectedStage(ctx context.Context, stageIDs []image.StageID, f func(ctx context.Context, stageID image.StageID) error) error {
	ids := append([]image.StageID(nil), stageIDs...)
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers:         m.MaxNumberOfWorkers(),
		InitDockerCLIForEachWorker: true,
	}, func(ctx context.Context, taskId int) error {
		return f(ctx, ids[taskId])
	})
}

func (m *StorageManager) ForEachDeleteStageCustomTag(ctx context.Context, ids []string, f func(ctx context.Context, tag string, err error) error) error {
	return parallel.DoTasks(ctx, len(ids), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		id := ids[taskId]

		// The custom-tag alias image lives where custom tags are published
		// (final repo if set, images repo otherwise); its metadata record
		// lives in the meta repo.
		if err := m.Storages.CustomTags().DeleteStageCustomTag(ctx, id); err != nil {
			return f(ctx, id, fmt.Errorf("unable to delete stage custom tag: %w", err))
		}
		if err := m.GetMetaStorage().UnregisterStageCustomTag(ctx, id); err != nil {
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
		metadata, err := m.GetMetaStorage().GetStageCustomTagMetadata(ctx, id)
		return f(ctx, id, metadata, err)
	})
}
