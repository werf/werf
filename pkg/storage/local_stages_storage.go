package storage

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	LocalStage_ImageRepoFormat              = "%s"
	LocalStage_ImageFormatWithCreationTs    = "%s:%s-%d"
	FilterReferenceLocalStageByDigestFormat = "%s:%s*"
	LocalStage_ImageFormat                  = "%s:%s"

	LocalImportMetadata_ImageNameFormat = "werf-import-metadata/%s"
	LocalImportMetadata_TagFormat       = "%s"

	ImageDeletionFailedDueToUsedByContainerErrorTip = "Use --force option to remove all containers that are based on deleting werf docker images"
)

func IsImageDeletionFailedDueToUsingByContainerErr(err error) bool {
	return strings.HasSuffix(err.Error(), ImageDeletionFailedDueToUsedByContainerErrorTip)
}

type LocalStagesStorage struct {
	ContainerBackend container_backend.ContainerBackend

	sequence atomic.Uint64

	imagesCacheMutex sync.Mutex
	imagesCache      map[string]localProjectSnapshot

	listingFlightMutex sync.Mutex
	listingFlights     map[string]*localProjectFlight
}

// localProjectSnapshot is an immutable record of the stage references of a single project: the
// listing that produced them and the references published by this process that no listing has
// confirmed yet. Only the references are kept, not the image summaries, so the listing payload of
// a daemon full of labeled images does not stay resident.
type localProjectSnapshot struct {
	initialized       bool
	references        []string
	listingStartedSeq uint64
	pushedReferences  map[string]uint64
}

type localProjectListing struct {
	references        []string
	listingStartedSeq uint64
}

// nextSequence orders requests, listings and publications of this process against each other
// without relying on the wall clock, whose resolution lets two of them share an instant.
func (storage *LocalStagesStorage) nextSequence() uint64 {
	return storage.sequence.Add(1)
}

// localProjectFlight is the listing of one project that is running right now. Which caller leads it
// and which ones join it is decided by whether this entry was found, under the flight mutex, so the
// roles are never inferred afterwards from a result a cancellation may have hidden. The entry is
// removed and done is closed in the same critical section, so a caller either joins a listing that
// will still answer it or starts the next one.
type localProjectFlight struct {
	done    chan struct{}
	listing localProjectListing
	err     error
}

func NewLocalStagesStorage(containerBackend container_backend.ContainerBackend) *LocalStagesStorage {
	return &LocalStagesStorage{ContainerBackend: containerBackend}
}

func (storage *LocalStagesStorage) FilterStageDescSetAndProcessRelatedData(ctx context.Context, stageDescSet image.StageDescSet, opts FilterStagesAndProcessRelatedDataOptions) (image.StageDescSet, error) {
	containersOpts := container_backend.ContainersOptions{}
	for stageDesc := range stageDescSet.Iter() {
		containersOpts.Filters = append(containersOpts.Filters, image.ContainerFilter{Ancestor: stageDesc.Info.ID})
	}
	containers, err := storage.ContainerBackend.Containers(ctx, containersOpts)
	if err != nil {
		return nil, err
	}

	stageDescSetToExclude := image.NewStageDescSet()
	var containerListToRemove []image.Container
	for _, container := range containers {
		for stageDesc := range stageDescSet.Iter() {
			imageInfo := stageDesc.Info

			if imageInfo.ID == container.ImageID {
				switch {
				case opts.SkipUsedImage:
					logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", imageInfo.LogName(), container.LogName())
					stageDescSetToExclude.Add(stageDesc)
				case opts.RmContainersThatUseImage:
					containerListToRemove = append(containerListToRemove, container)
				default:
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", imageInfo.LogName(), container.LogName(), ImageDeletionFailedDueToUsedByContainerErrorTip)
				}
			}
		}
	}

	if err := storage.deleteContainers(ctx, containerListToRemove, opts.RmForce); err != nil {
		return nil, err
	}

	return stageDescSet.Difference(stageDescSetToExclude), nil
}

func (storage *LocalStagesStorage) deleteContainers(ctx context.Context, containers []image.Container, rmForce bool) error {
	for _, container := range containers {
		if err := storage.ContainerBackend.Rm(ctx, container.ID, container_backend.RmOpts{Force: rmForce}); err != nil {
			return fmt.Errorf("unable to remove container %q: %w", container.ID, err)
		}
	}
	return nil
}

func (storage *LocalStagesStorage) GetStagesIDs(ctx context.Context, projectName string, opts ...Option) ([]image.StageID, error) {
	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName)))
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("label", fmt.Sprintf("%s=%s", image.WerfLabel, projectName)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}
	return images.ConvertToStages()
}

func (storage *LocalStagesStorage) GetStagesIDsByDigest(ctx context.Context, projectName, digest string, parentStageCreationTs int64, opts ...Option) ([]image.StageID, error) {
	var requestSeq uint64
	if !makeOptions(opts...).withCache {
		// Recorded before anything else, so that a listing accepted by this call also covers every
		// lock its caller already holds.
		requestSeq = storage.sequence.Load()
	} else if cached, isCached := storage.loadProjectSnapshot(projectName); isCached {
		return selectProjectStages(ctx, cached.references, projectName, digest, parentStageCreationTs)
	}

	listing, _, err := storage.refreshProjectListing(ctx, projectName, requestSeq)
	if err != nil {
		return nil, err
	}

	return selectProjectStages(ctx, listing.references, projectName, digest, parentStageCreationTs)
}

func selectProjectStages(ctx context.Context, references []string, projectName, digest string, parentStageCreationTs int64) ([]image.StageID, error) {
	matched := lo.Filter(references, func(reference string, _ int) bool {
		return strings.HasPrefix(reference, fmt.Sprintf(LocalStage_ImageFormat, projectName, digest))
	})
	if len(matched) == 0 {
		return nil, nil
	}

	stagesIDs, err := image.ImagesList{{RepoTags: matched}}.ConvertToStages()
	if err != nil {
		return nil, fmt.Errorf("unable to convert images to stages: %w", err)
	}

	var resultStageIDs []image.StageID
	for _, stageID := range stagesIDs {
		if parentStageCreationTs > stageID.CreationTs {
			logboek.Context(ctx).Debug().LogF("Skip stage %s (parent stage creation timestamp %d is greater than the stage creation timestamp %d)\n", stageID.String(), parentStageCreationTs, stageID.CreationTs)
			continue
		}

		resultStageIDs = append(resultStageIDs, stageID)
	}

	return resultStageIDs, nil
}

func (storage *LocalStagesStorage) refreshProjectListing(ctx context.Context, projectName string, requestSeq uint64) (localProjectListing, bool, error) {
	var joined bool
	for {
		listing, joinedNow, err := storage.waitProjectListing(ctx, projectName)
		joined = joined || joinedNow
		if err != nil {
			return localProjectListing{}, joined, err
		}
		// The listing this call joined had already snapshotted the daemon before the caller took the
		// locks the result has to cover, so it is unusable. The next listing can only start after
		// this one finished, so waiting for it never runs two listings in parallel.
		if listing.listingStartedSeq <= requestSeq {
			continue
		}
		return listing, joined, nil
	}
}

func (storage *LocalStagesStorage) waitProjectListing(ctx context.Context, projectName string) (localProjectListing, bool, error) {
	if err := ctx.Err(); err != nil {
		return localProjectListing{}, false, err
	}

	storage.listingFlightMutex.Lock()
	flight, joined := storage.listingFlights[projectName]
	if !joined {
		flight = &localProjectFlight{done: make(chan struct{})}
		if storage.listingFlights == nil {
			storage.listingFlights = make(map[string]*localProjectFlight)
		}
		storage.listingFlights[projectName] = flight
		// The listing runs on the context of the caller that started it, so canceling that caller
		// still aborts the backend call, while a joiner giving up leaves the listing running for
		// everyone else waiting on it.
		go storage.runProjectListing(ctx, projectName, flight)
	}
	storage.listingFlightMutex.Unlock()

	select {
	case <-ctx.Done():
		return localProjectListing{}, joined, ctx.Err()
	case <-flight.done:
		return flight.listing, joined, flight.err
	}
}

func (storage *LocalStagesStorage) runProjectListing(ctx context.Context, projectName string, flight *localProjectFlight) {
	listingStartedSeq := storage.nextSequence()
	storage.registerProjectListing(projectName)

	var listing localProjectListing
	images, err := storage.listImages(ctx, fmt.Sprintf(LocalStage_ImageRepoFormat, projectName))
	if err == nil {
		listing = localProjectListing{
			references:        storage.storeProjectSnapshot(projectName, localStageReferences(images), listingStartedSeq),
			listingStartedSeq: listingStartedSeq,
		}
	}

	storage.listingFlightMutex.Lock()
	flight.listing, flight.err = listing, err
	delete(storage.listingFlights, projectName)
	close(flight.done)
	storage.listingFlightMutex.Unlock()
}

func (storage *LocalStagesStorage) registerProjectListing(projectName string) {
	storage.imagesCacheMutex.Lock()
	defer storage.imagesCacheMutex.Unlock()

	storage.putProjectSnapshot(projectName, storage.imagesCache[projectName])
}

func (storage *LocalStagesStorage) storeProjectSnapshot(projectName string, references []string, listingStartedSeq uint64) []string {
	storage.imagesCacheMutex.Lock()
	defer storage.imagesCacheMutex.Unlock()

	entry := storage.imagesCache[projectName]
	pushedReferences := map[string]uint64{}
	for reference, pushedSeq := range entry.pushedReferences {
		// A listing started after the publication is authoritative and drops the reference, so a
		// stage removed from the daemon cannot stay cached forever.
		if pushedSeq <= listingStartedSeq || slices.Contains(references, reference) {
			continue
		}
		references = append(references, reference)
		pushedReferences[reference] = pushedSeq
	}

	// A listing that started earlier carries an older snapshot of the daemon, whatever the order of
	// completion: it must not drop references a later listing already confirmed.
	if entry.initialized && entry.listingStartedSeq > listingStartedSeq {
		return references
	}

	storage.putProjectSnapshot(projectName, localProjectSnapshot{
		initialized:       true,
		references:        references,
		listingStartedSeq: listingStartedSeq,
		pushedReferences:  pushedReferences,
	})
	return references
}

func (storage *LocalStagesStorage) loadProjectSnapshot(projectName string) (localProjectSnapshot, bool) {
	storage.imagesCacheMutex.Lock()
	defer storage.imagesCacheMutex.Unlock()

	entry, isCached := storage.imagesCache[projectName]
	return entry, isCached && entry.initialized
}

func (storage *LocalStagesStorage) putProjectSnapshot(projectName string, entry localProjectSnapshot) {
	if storage.imagesCache == nil {
		storage.imagesCache = make(map[string]localProjectSnapshot)
	}
	storage.imagesCache[projectName] = entry
}

func (storage *LocalStagesStorage) listImages(ctx context.Context, reference string) (image.ImagesList, error) {
	images, err := storage.ContainerBackend.Images(ctx, container_backend.ImagesOptions{
		Filters: []util.Pair[string, string]{util.NewPair("reference", reference)},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %w", err)
	}
	return images, nil
}

// localStageReferences flattens an image listing into the references ConvertToStages needs,
// dropping the "localhost/" prefix Buildah reports so that a reference is comparable whichever
// backend produced it.
func localStageReferences(images image.ImagesList) []string {
	return lo.FlatMap(images, func(summary image.Summary, _ int) []string {
		return lo.Map(summary.RepoTags, func(repoTag string, _ int) string {
			return trimLocalStageReference(repoTag)
		})
	})
}

func trimLocalStageReference(reference string) string {
	return strings.TrimPrefix(reference, "localhost/")
}

func (storage *LocalStagesStorage) rememberPublishedStage(reference string) {
	reference = trimLocalStageReference(reference)
	projectName, tag := image.ParseRepositoryAndTag(reference)
	if projectName == "" || tag == "" {
		return
	}

	storage.imagesCacheMutex.Lock()
	defer storage.imagesCacheMutex.Unlock()

	entry, isCached := storage.imagesCache[projectName]
	if !isCached {
		return
	}

	pushedReferences := maps.Clone(entry.pushedReferences)
	if pushedReferences == nil {
		pushedReferences = map[string]uint64{}
	}
	// Recorded even when the reference is already listed: a listing in flight may have snapshotted
	// the daemon without it.
	pushedReferences[reference] = storage.nextSequence()
	entry.pushedReferences = pushedReferences

	if entry.initialized && !slices.Contains(entry.references, reference) {
		entry.references = append(slices.Clone(entry.references), reference)
	}

	storage.putProjectSnapshot(projectName, entry)
}

func (storage *LocalStagesStorage) GetStageDesc(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDesc, error) {
	stageImageName := storage.ConstructStageImageName(projectName, stageID.Digest, stageID.CreationTs)
	info, err := storage.ContainerBackend.GetImageInfo(ctx, stageImageName, container_backend.GetImageInfoOpts{})
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s info: %w", stageImageName, err)
	}

	if info == nil {
		return nil, ErrStageNotFound
	}

	return &image.StageDesc{
		StageID: image.NewStageID(stageID.Digest, stageID.CreationTs),
		Info:    info,
	}, nil
}

func (storage *LocalStagesStorage) ExportStage(ctx context.Context, stageDesc *image.StageDesc, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error {
	if err := storage.ContainerBackend.Tag(ctx, stageDesc.Info.Name, destinationReference, container_backend.TagOpts{}); err != nil {
		return fmt.Errorf("unable to tag %q as %q: %w", stageDesc.Info.Name, destinationReference, err)
	}
	defer func() {
		_ = storage.ContainerBackend.Rmi(ctx, destinationReference, container_backend.RmiOpts{Force: true})
	}()

	if err := storage.ContainerBackend.Push(ctx, destinationReference, container_backend.PushOpts{}); err != nil {
		return fmt.Errorf("unable to push %q: %w", destinationReference, err)
	}
	return docker_registry.API().MutateAndPushImage(ctx, destinationReference, destinationReference, api.WithConfigMutation(mutateExportStageConfig(mutateConfigFunc)))
}

func (storage *LocalStagesStorage) DeleteStage(ctx context.Context, stageDesc *image.StageDesc, options DeleteImageOptions) error {
	var imageReferences []string
	imageInfo := stageDesc.Info

	if imageInfo.Name == "" {
		imageReferences = append(imageReferences, imageInfo.ID)
	} else {
		isDanglingImage := imageInfo.Name == "<none>:<none>"
		isTaglessImage := !isDanglingImage && imageInfo.Tag == "<none>"

		if isDanglingImage || isTaglessImage {
			imageReferences = append(imageReferences, imageInfo.ID)
		} else {
			imageReferences = append(imageReferences, imageInfo.Name)
		}
	}

	for _, ref := range imageReferences {
		if err := storage.ContainerBackend.Rmi(ctx, ref, container_backend.RmiOpts{Force: options.RmiForce}); err != nil {
			return fmt.Errorf("unable to remove %q: %w", ref, err)
		}
	}
	return nil
}

func (storage *LocalStagesStorage) AddStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) CheckStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) DeleteStageCustomTag(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RejectStage(_ context.Context, _, _ string, _ int64) error {
	return nil
}

func (storage *LocalStagesStorage) GetRejectedStageIDs(_ context.Context, _ ...Option) ([]image.StageID, error) {
	return nil, nil
}

func (storage *LocalStagesStorage) DeleteRejectedStageImage(_ context.Context, _ image.StageID, _ DeleteImageOptions) error {
	return nil
}

func (storage *LocalStagesStorage) DeleteRejectedStageRecord(_ context.Context, _ image.StageID, _ DeleteImageOptions) error {
	return nil
}

func (storage *LocalStagesStorage) ConstructStageImageName(projectName, digest string, creationTs int64) string {
	if creationTs == 0 {
		return fmt.Sprintf(LocalStage_ImageFormat, projectName, digest)
	}
	return fmt.Sprintf(LocalStage_ImageFormatWithCreationTs, projectName, digest, creationTs)
}

func (storage *LocalStagesStorage) FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return nil
}

func (storage *LocalStagesStorage) StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	if err := storage.ContainerBackend.TagImageByName(ctx, img); err != nil {
		return err
	}
	storage.rememberPublishedStage(img.Name())
	return nil
}

func (storage *LocalStagesStorage) ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) CreateRepo(ctx context.Context) error { return nil }

func (storage *LocalStagesStorage) DeleteRepo(ctx context.Context) error { return nil }

func (storage *LocalStagesStorage) AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalStagesStorage) RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalStagesStorage) IsManagedImageExist(ctx context.Context, projectName, imageNameOrManagedImageName string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	return []string{}, nil
}

func (storage *LocalStagesStorage) PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error {
	return nil
}

func (storage *LocalStagesStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error {
	return nil
}

func (storage *LocalStagesStorage) IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return map[string]map[string][]string{}, map[string]map[string][]string{}, nil
}

func (storage *LocalStagesStorage) GetImportMetadata(ctx context.Context, projectName, id string) (*ImportMetadata, error) {
	if debugStagesStorage() {
		logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetImportMetadata %s %s\n", projectName, id)
	}

	fullImageName := makeLocalImportMetadataName(projectName, id)

	info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{})
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s info: %w", fullImageName, err)
	}
	if info == nil {
		return nil, ErrImportMetadataNotFound
	}
	return newImportMetadataFromLabels(info.Labels), nil
}

func (storage *LocalStagesStorage) PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata, opts PutImportMetadataOptions) error {
	if debugStagesStorage() {
		logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.PutImportMetadata %s %v\n", projectName, metadata)
	}

	fullImageName := makeLocalImportMetadataName(projectName, metadata.ImportSourceID)

	if !opts.Force {
		if info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{}); err != nil {
			return fmt.Errorf("unable to check existence of image %s: %w", fullImageName, err)
		} else if info != nil {
			return nil
		}
	}

	labels := metadata.ToLabels()
	labels = append(labels, fmt.Sprintf("%s=%s", image.WerfLabel, projectName))
	if err := storage.ContainerBackend.PostManifest(ctx, fullImageName, container_backend.PostManifestOpts{Labels: labels}); err != nil {
		return fmt.Errorf("unable to post manifest %q: %w", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) RmImportMetadata(ctx context.Context, projectName, id string) error {
	if debugStagesStorage() {
		logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.RmImportMetadata %s %s\n", projectName, id)
	}

	fullImageName := makeLocalImportMetadataName(projectName, id)

	if info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to check existence of image %s: %w", fullImageName, err)
	} else if info != nil {
		return nil
	}

	if err := storage.ContainerBackend.Rmi(ctx, fullImageName, container_backend.RmiOpts{Force: true}); err != nil {
		return fmt.Errorf("unable to remove image %s: %w", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) GetImportMetadataIDs(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	if debugStagesStorage() {
		logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetImportMetadataIDs %s\n", projectName)
	}

	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName)))
	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}

	var tags []string
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			_, tag := image.ParseRepositoryAndTag(repoTag)
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (storage *LocalStagesStorage) GetClientIDRecords(_ context.Context, _ string, _ ...Option) ([]*ClientIDRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostClientIDRecord(_ context.Context, _ string, _ *ClientIDRecord) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostMultiplatformImage(_ context.Context, _, _ string, _ []*image.Info, _ []string) error {
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) Address() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...Option) ([]string, error) {
	return nil, nil
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadata(_ context.Context, _ string) (*CustomTagMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RegisterStageCustomTag(_ context.Context, _ string, _ *image.StageDesc, tag string) error {
	return nil
}

func (storage *LocalStagesStorage) UnregisterStageCustomTag(_ context.Context, _ string) error {
	return nil
}

func (storage *LocalStagesStorage) CopyFromStorage(_ context.Context, _ StagesStorage, _ string, _ image.StageID, _ CopyFromStorageOptions) (*image.StageDesc, error) {
	panic("not implemented")
}

func makeLocalImportMetadataName(projectName, importSourceID string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName),
			fmt.Sprintf(LocalImportMetadata_TagFormat, importSourceID),
		}, ":",
	)
}

func (storage *LocalStagesStorage) GetSyncServerRecords(ctx context.Context, projectName string, opts ...Option) ([]*SyncServerRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostSyncServerRecord(ctx context.Context, projectName string, rec *SyncServerRecord) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) GetLastCleanupRecord(ctx context.Context, projectName string, opts ...Option) (*CleanupRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostLastCleanupRecord(ctx context.Context, projectName string) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostManifest(ctx context.Context, ref string, opts container_backend.PostManifestOpts) error {
	if err := storage.ContainerBackend.PostManifest(ctx, ref, opts); err != nil {
		return fmt.Errorf("unable to post manifest %s: %w", ref, err)
	}

	return nil
}

func (storage *LocalStagesStorage) MutateAndPushImage(ctx context.Context, src, _ string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error {
	if debugImageSpec() {
		if err := logboek.Context(ctx).Debug().LogBlock("-- LocalStagesStorage.MutateAndPushImage imageSpecConfig").DoError(func() error {
			newConfigData, err := yaml.Marshal(newConfig)
			if err != nil {
				return fmt.Errorf("unable to yaml marshal new config: %w", err)
			}

			logboek.Context(ctx).Debug().LogF(string(newConfigData))
			return nil
		}); err != nil {
			return err
		}
	}

	newId, err := container_backend.MutateAndPushImage(ctx, src, stageImage.GetTargetPlatform(), newConfig, storage.ContainerBackend)
	if err != nil {
		return err
	}

	stageImage.SetBuiltID(newId)

	if err := storage.ContainerBackend.TagImageByName(ctx, stageImage); err != nil {
		return fmt.Errorf("unable to tag image %q: %w", stageImage.Name(), err)
	}
	storage.rememberPublishedStage(stageImage.Name())

	return nil
}
