package host_cleaning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
)

var ErrUnsupportedContainerBackend = errors.New("unsupported container backend")

var errOptionDryRunNotSupported = errors.New("option dry-run not supported")

type RunGCOptions struct {
	AllowedStorageVolumeUsageBytes       uint64
	AllowedStorageVolumeUsageMarginBytes uint64
	StoragePath                          string
	Force                                bool
	DryRun                               bool
}

type RunAutoGCOptions struct {
	AllowedStorageVolumeUsageBytes uint64
	StoragePath                    string
}

//go:generate mockgen -package mock -destination ../../test/mock/locker.go github.com/werf/lockgate Locker

type LocalBackendCleaner struct {
	backend     container_backend.ContainerBackend
	backendType containerBackendType
	locker      lockgate.Locker
	// refs for stubbing in testing
	volumeutilsGetVolumeUsageByPath func(ctx context.Context, path string) (volumeutils.VolumeUsage, error)
	werfGetWerfLastRunAtV1_1        func(ctx context.Context) (time.Time, error)
	lrumetaGetImageLastAccessTime   func(ctx context.Context, imageRef string) (time.Time, error)
}

func NewLocalBackendCleaner(backend container_backend.ContainerBackend, locker lockgate.Locker) (*LocalBackendCleaner, error) {
	cleaner := &LocalBackendCleaner{
		backend: backend,
		locker:  locker,
		// refs for stubbing in testing
		volumeutilsGetVolumeUsageByPath: volumeutils.GetVolumeUsageByPath,
		werfGetWerfLastRunAtV1_1:        werf.GetWerfLastRunAtV1_1,
		lrumetaGetImageLastAccessTime:   lrumeta.CommonLRUImagesCache.GetImageLastAccessTime,
	}

	backendType, err := resolveContainerBackendType(backend)
	if err != nil {
		return cleaner, err
	}
	cleaner.backendType = backendType

	return cleaner, nil
}

func (cleaner *LocalBackendCleaner) BackendName() string {
	return cleaner.backendType.String()
}

func (cleaner *LocalBackendCleaner) backendStoragePath(ctx context.Context, storagePath string) (string, error) {
	backendStoragePath := storagePath

	if backendStoragePath == "" {
		info, err := cleaner.backend.Info(ctx)
		if err != nil {
			return "", fmt.Errorf("error getting local %s backend info: %w", cleaner.BackendName(), err)
		}
		backendStoragePath = info.StoreGraphRoot
	}

	// assert path existence and permissions
	if _, err := os.Stat(backendStoragePath); err != nil {
		return "", fmt.Errorf("error accessing %q: %w", backendStoragePath, err)
	}

	return backendStoragePath, nil
}

func (cleaner *LocalBackendCleaner) ShouldRunAutoGC(ctx context.Context, options RunAutoGCOptions) (bool, error) {
	backendStoragePath, err := cleaner.backendStoragePath(ctx, options.StoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting local %s backend storage path: %w", cleaner.BackendName(), err)
	}

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", backendStoragePath, err)
	}
	return vu.UsedBytes > options.AllowedStorageVolumeUsageBytes, nil
}

// werfImages returns werf images are safe for removing
func (cleaner *LocalBackendCleaner) werfImages(ctx context.Context) (image.ImagesList, error) {
	images0, err := cleaner.werfImagesByLabels(ctx)
	if err != nil {
		return nil, err
	}
	images1, err := cleaner.werfImagesByLegacyLabels(ctx)
	if err != nil {
		return nil, err
	}
	images2, err := cleaner.werfImagesByLastRun(ctx)
	if err != nil {
		return nil, err
	}

	images := slices.Concat(images0, images1, images2)

	lastUsedAtMap := make(map[string]time.Time, len(images))

	for _, img := range images {
		data, _ := json.Marshal(img)
		logboek.Context(ctx).Debug().LogF("Image summary:\n%s\n---\n", data)

		lastUsedAtMap[img.ID], err = cleaner.maxLastUsedAtForImage(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("unable to max last used at for image: %w", err)
		}
	}

	slices.SortFunc(images, func(a, b image.Summary) int {
		aTime := lastUsedAtMap[a.ID]
		bTime := lastUsedAtMap[b.ID]
		return aTime.Compare(bTime)
	})

	return images, nil
}

func (cleaner *LocalBackendCleaner) werfImagesByLabels(ctx context.Context) (image.ImagesList, error) {
	list, err := cleaner.backend.Images(ctx, buildImagesOptions(
		filter.DanglingFalse.ToPair(),
		util.NewPair("label", image.WerfLabel),
	))
	if err != nil {
		return nil, fmt.Errorf("unable to get werf %s images: %w", cleaner.BackendName(), err)
	}

	return list, nil
}

func (cleaner *LocalBackendCleaner) werfImagesByLegacyLabels(ctx context.Context) (image.ImagesList, error) {
	// Process legacy v1.1 images
	list, err := cleaner.backend.Images(ctx, buildImagesOptions(
		filter.DanglingFalse.ToPair(),
		util.NewPair("label", image.WerfLabel),
		util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
	))
	if err != nil {
		return nil, fmt.Errorf("unable to get werf v1.1 legacy %s images: %w", cleaner.BackendName(), err)
	}

	return list, nil
}

func (cleaner *LocalBackendCleaner) werfImagesByLastRun(ctx context.Context) (image.ImagesList, error) {
	// **NOTICE** Remove v1.1 last-run-at timestamp check when v1.1 reaches its end of life

	t, err := cleaner.werfGetWerfLastRunAtV1_1(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting v1.1 last run timestamp: %w", err)
	}

	var images image.ImagesList

	// No werf v1.1 runs on this host.
	// This is stupid check, but the only available safe option at the moment.
	if t.IsZero() {
		list, err := cleaner.backend.Images(ctx, buildImagesOptions(
			filter.DanglingFalse.ToPair(),

			util.NewPair("reference", "*client-id-*"),
			util.NewPair("reference", "*managed-image-*"),
			util.NewPair("reference", "*meta-*"),
			util.NewPair("reference", "*-rejected"),

			util.NewPair("reference", "werf-client-id/*"),
			util.NewPair("reference", "werf-managed-images/*"),
			util.NewPair("reference", "werf-images-metadata-by-commit/*"),
		))
		if err != nil {
			return nil, fmt.Errorf("unable to get werf service images: %w", err)
		}

		images = slices.Grow(images, len(list))

		for _, img := range list {
			// **NOTICE.** Cannot remove by werf label, because currently there is no such label for service-images by historical reasons.
			// So check by size at least for now.
			if img.Size != 0 {
				continue
			}

			images = append(images, img)
		}
	}

	return images, err
}

func (cleaner *LocalBackendCleaner) maxLastUsedAtForImage(ctx context.Context, img image.Summary) (time.Time, error) {
	lastUsedAt := img.Created

	for _, ref := range img.RepoTags {
		lastRecentlyUsedAt, err := cleaner.lrumetaGetImageLastAccessTime(ctx, ref)
		if err != nil {
			return time.Time{}, fmt.Errorf("error accessing last recently used images cache: %w", err)
		}

		if lastUsedAt.Before(lastRecentlyUsedAt) {
			lastUsedAt = lastRecentlyUsedAt
		}
	}

	return lastUsedAt, nil
}

func (cleaner *LocalBackendCleaner) RunGC(ctx context.Context, options RunGCOptions) error {
	backendStoragePath, err := cleaner.backendStoragePath(ctx, options.StoragePath)
	if err != nil {
		return fmt.Errorf("error getting local %s backend storage path: %w", cleaner.BackendName(), err)
	}

	// We can clarify StoragePath from now to further usage
	options.StoragePath = backendStoragePath

	logboek.Context(ctx).LogF("Storage path: %s\n", options.StoragePath)
	logboek.Context(ctx).LogOptionalLn()

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %w", backendStoragePath, err)
	}

	// initialVolumeUsage stores the baseline volume usage to calculate the total factual freed space at the end of the GC process.
	initialVolumeUsage := vu

	if vu.UsedBytes <= options.AllowedStorageVolumeUsageBytes {
		logboek.Context(ctx).LogBlock("Check storage").Do(func() {
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed volume usage: %s <= %s — %s\n", utils.GreenF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.BlueF("%s (%.2f%%)", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), vu.BytesToPercentage(options.AllowedStorageVolumeUsageBytes)), utils.GreenF("OK"))
		})

		return nil
	}

	targetVolumeUsageBytes := uint64(math.Max(float64(options.AllowedStorageVolumeUsageBytes)-float64(options.AllowedStorageVolumeUsageMarginBytes), 0))

	logboek.Context(ctx).LogBlock("Check storage").Do(func() {
		neededToFreeBytes := uint64(math.Max(float64(vu.UsedBytes)-float64(targetVolumeUsageBytes), 0))
		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Allowed level exceeded: %s > %s — %s\n", utils.RedF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.YellowF("%s (%.2f%%)", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), vu.BytesToPercentage(options.AllowedStorageVolumeUsageBytes)), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Target level after cleanup: %s - %s (margin) = %s\n", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), humanize.Bytes(options.AllowedStorageVolumeUsageMarginBytes), utils.BlueF("%s (%.2f%%)", humanize.Bytes(targetVolumeUsageBytes), vu.BytesToPercentage(targetVolumeUsageBytes)))
		logboek.Context(ctx).LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(neededToFreeBytes)))
	})

	// Step 1. Prune unused anonymous volumes
	err = logboek.Context(ctx).LogBlock("Prune all unused anonymous volumes").DoError(func() error {
		reportVolumes, err := cleaner.pruneVolumes(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		var spaceReclaimed uint64
		if spaceReclaimed, vu, err = cleaner.measureReclaimedSpace(ctx, options.StoragePath, vu); err != nil {
			return err
		}
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(spaceReclaimed)))
		logDeletedItems(ctx, reportVolumes.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune unused anonymous volumes: %w", err)
	}

	// Step 2. Prune werf dangling images
	err = logboek.Context(ctx).LogBlock("Prune werf dangling images created more than 1 hour ago").DoError(func() error {
		reportImages, err := cleaner.pruneImages(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		var spaceReclaimed uint64
		if spaceReclaimed, vu, err = cleaner.measureReclaimedSpace(ctx, options.StoragePath, vu); err != nil {
			return err
		}
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(spaceReclaimed)))
		logDeletedItems(ctx, reportImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune werf dangling images: %w", err)
	}

	if vu.UsedBytes <= options.AllowedStorageVolumeUsageBytes {
		logboek.Context(ctx).LogBlock("Check storage").Do(func() {
			totalFreedBytes := uint64(math.Max(float64(initialVolumeUsage.UsedBytes)-float64(vu.UsedBytes), 0))
			logboek.Context(ctx).LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(totalFreedBytes)))
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed level exceeded: %s > %s — %s\n", utils.RedF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.YellowF("%s (%.2f%%)", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), vu.BytesToPercentage(options.AllowedStorageVolumeUsageBytes)), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).LogF("Target level after cleanup: %s - %s (margin) = %s\n", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), humanize.Bytes(options.AllowedStorageVolumeUsageMarginBytes), utils.BlueF("%s (%.2f%%)", humanize.Bytes(targetVolumeUsageBytes), vu.BytesToPercentage(targetVolumeUsageBytes)))
		})

		return nil
	}

	// Step 3. Remove werf containers
	err = logboek.Context(ctx).LogBlock("Cleanup werf containers").DoError(func() error {
		reportWerfContainers, err := cleaner.cleanupWerfContainers(ctx, options)
		if err != nil {
			return err
		}

		var spaceReclaimed uint64
		if spaceReclaimed, vu, err = cleaner.measureReclaimedSpace(ctx, options.StoragePath, vu); err != nil {
			return err
		}
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(spaceReclaimed)))
		logDeletedItems(ctx, reportWerfContainers.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to remove werf containers: %w", err)
	}

	// Step 4. Remove werf images
	err = logboek.Context(ctx).LogBlock("Cleanup werf images").DoError(func() error {
		reportWerfImages, err := cleaner.cleanupWerfImages(ctx, options, targetVolumeUsageBytes)
		if err != nil {
			return err
		}

		var spaceReclaimed uint64
		if spaceReclaimed, vu, err = cleaner.measureReclaimedSpace(ctx, options.StoragePath, vu); err != nil {
			return err
		}
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(spaceReclaimed)))
		logDeletedItems(ctx, reportWerfImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to cleanup werf images: %w", err)
	}

	logboek.Context(ctx).LogBlock("Check storage").Do(func() {
		totalFreedBytes := uint64(math.Max(float64(initialVolumeUsage.UsedBytes)-float64(vu.UsedBytes), 0))
		logboek.Context(ctx).LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(totalFreedBytes)))
		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Allowed level exceeded: %s > %s — %s\n", utils.RedF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.YellowF("%s (%.2f%%)", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), vu.BytesToPercentage(options.AllowedStorageVolumeUsageBytes)), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Target level after cleanup: %s - %s (margin) = %s\n", humanize.Bytes(options.AllowedStorageVolumeUsageBytes), humanize.Bytes(options.AllowedStorageVolumeUsageMarginBytes), utils.BlueF("%s (%.2f%%)", humanize.Bytes(targetVolumeUsageBytes), vu.BytesToPercentage(targetVolumeUsageBytes)))
	})

	if vu.UsedBytes > targetVolumeUsageBytes {
		logboek.Context(ctx).Info().LogOptionalLn()
		logboek.Context(ctx).Info().LogF("NOTE: Detected high %s storage volume usage, while no werf images available to cleanup!\n", cleaner.BackendName())
		logboek.Context(ctx).Info().LogF("NOTE:\n")
		logboek.Context(ctx).Info().LogF("NOTE: werf tries to maintain host clean by deleting:\n")
		logboek.Context(ctx).Info().LogF("NOTE:  - old unused files from werf caches (which are stored in the ~/.werf/local_cache);\n")
		logboek.Context(ctx).Info().LogF("NOTE:  - old temporary service files /tmp/werf-project-data-* and /tmp/werf-config-render-*;\n")
		logboek.Context(ctx).Info().LogF("NOTE:  - least recently used werf images except local stages storage images (images built with 'werf build' without '--repo' param, or with '--stages-storage=:local' param for the werf v1.1).\n")
		logboek.Context(ctx).Info().LogOptionalLn()
	}

	return nil
}

// measureReclaimedSpace gets the actual disk state, calculates the factual
// freed space relative to vuBefore and returns this volume along with the new state.
func (cleaner *LocalBackendCleaner) measureReclaimedSpace(ctx context.Context, storagePath string, vuBefore volumeutils.VolumeUsage) (uint64, volumeutils.VolumeUsage, error) {
	vuAfter, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, storagePath)
	if err != nil {
		return 0, volumeutils.VolumeUsage{}, fmt.Errorf("error getting volume usage by path %q: %w", storagePath, err)
	}

	spaceReclaimed := uint64(math.Max(float64(vuBefore.UsedBytes)-float64(vuAfter.UsedBytes), 0))

	return spaceReclaimed, vuAfter, nil
}

// pruneImages removes werf dangling images
func (cleaner *LocalBackendCleaner) pruneImages(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	filters := filter.FilterList{
		// 1. Select all dangling images.
		filter.DanglingTrue,
		// 2. From all dangling images select only werf's dangling images.
		filter.NewFilter("label", image.WerfLabel),
		// 3. From werf's dangling images select only images which were created more than 15 minutes ago.
		// Explanation: in Stapel mode werf relies on a "dangling" image for some time before tagging its image.
		filter.NewFilter("until", "15m"),

		// Both backends support filters listed above:
		// Docker: https://github.com/moby/moby/blob/25.0/daemon/containerd/image_prune.go#L22
		// Buildah: https://github.com/containers/common/blob/v0.58/libimage/filters.go#L111
	}

	if options.DryRun {
		list, err := cleaner.backend.Images(ctx, buildImagesOptions(filters.ToPairs()...))
		if err != nil {
			return cleanupReport{}, err
		}
		return mapImageListToCleanupReport(list), nil
	}

	report, err := cleaner.backend.PruneImages(ctx, prune.Options{Filters: filters})
	switch {
	case errors.Is(err, container_backend.ErrImageUsedByContainer),
		errors.Is(err, container_backend.ErrPruneIsAlreadyRunning):
		logboek.Context(ctx).Info().LogF("NOTE: Ignore image pruning: %s\n", err.Error())
		return cleanupReport{}, nil
	case err != nil:
		return cleanupReport{}, err
	}

	return mapPruneReportToCleanupReport(report), err
}

// pruneVolumes removes all anonymous volumes not used by at least one container
func (cleaner *LocalBackendCleaner) pruneVolumes(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		// NOTE: Buildah does not give us a way to precalculate pruned size.
		// NOTE: Docker does not give us a way to precalculate pruned size.
		return cleanupReport{}, errOptionDryRunNotSupported
	}

	report, err := cleaner.backend.PruneVolumes(ctx, prune.Options{})

	switch {
	case errors.Is(err, container_backend.ErrPruneIsAlreadyRunning):
		logboek.Context(ctx).Info().LogF("NOTE: Ignore volume pruning: %s\n", err.Error())
		return cleanupReport{}, nil
	case err != nil:
		return cleanupReport{}, err
	}

	return mapPruneReportToCleanupReport(report), err
}

func (cleaner *LocalBackendCleaner) cleanupWerfContainers(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	containers, err := werfContainersByContainersOptions(ctx, cleaner.backend, buildContainersOptions())
	if err != nil {
		return cleanupReport{}, fmt.Errorf("cannot get build containers: %w", err)
	}

	report := cleanupReport{
		ItemsDeleted: make([]string, 0, len(containers)),
	}

	for _, container := range containers {
		containerName := werfContainerName(container)

		if containerName == "" {
			logboek.Context(ctx).Info().LogF("Ignore bad container %s\n", container.ID)
			continue
		}

		if ok, err := cleaner.isLocked(container_backend.ContainerLockName(containerName)); err != nil {
			return cleanupReport{}, fmt.Errorf("checking lock %q: %w", container_backend.ContainerLockName(containerName), err)
		} else if ok {
			continue
		}

		if err = cleaner.removeContainerRef(ctx, container.ID, options); err != nil {
			switch {
			case errors.Is(err, container_backend.ErrCannotRemovePausedContainer):
				logboek.Context(ctx).Info().LogF("Ignore paused container %s\n", logContainerName(container))
				continue
			case errors.Is(err, container_backend.ErrCannotRemoveRunningContainer):
				logboek.Context(ctx).Info().LogF("Ignore running container %s\n", logContainerName(container))
				continue
			case err != nil:
				logboek.Context(ctx).Info().LogF("Cannot remove container by id %q: %s\n", container.ID, err)
				continue
			}
		}

		report.ItemsDeleted = append(report.ItemsDeleted, container.ID)
	}

	return report.Normalize(), nil
}

func (cleaner *LocalBackendCleaner) cleanupWerfImages(ctx context.Context, options RunGCOptions, targetVolumeUsageBytes uint64) (cleanupReport, error) {
	images, err := cleaner.werfImages(ctx)
	if err != nil {
		return cleanupReport{}, err
	}

	report := cleanupReport{
		ItemsDeleted: make([]string, 0, len(images)),
	}

	// 1. Start the cleaning loop.
	startIndex := 0
	for {
		// 2. Re-calculate actual volume usage before deletion batch.
		vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return report, fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}

		// 3. Calculate how many bytes we still need to free.
		if vu.UsedBytes <= targetVolumeUsageBytes {
			break
		}
		bytesToFree := vu.UsedBytes - targetVolumeUsageBytes

		// 4. Find the next batch of images to delete based on the current disk state.
		// Returns -1 if the target is reached or the end of the list is encountered.
		n := countImagesToFree(images, startIndex, bytesToFree)
		// 5. Exit if target reached (bytesToFree == 0) or no more images to process.
		if n == -1 {
			break
		}

		// 6. Delete the identified batch of images.
		for _, imgSummary := range images[startIndex : n+1] {
			var ok bool
			var err error

			if len(imgSummary.RepoTags) > 0 {
				ok, err = cleaner.removeImageByRepoTags(ctx, options, imgSummary)
				if err != nil {
					return report, err
				}
			} else if len(imgSummary.RepoDigests) > 0 {
				ok, err = cleaner.removeImageByRepoDigests(ctx, options, imgSummary)
				if err != nil {
					return report, err
				}
			}

			if ok {
				report.ItemsDeleted = append(report.ItemsDeleted, imgSummary.ID)
			}
		}

		// 7. Re-calculate actual volume usage after deletion.
		// This is an expensive operation, so we do it once per batch.
		vuAfter, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return report, fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}
		// 8. If no space was reclaimed (e.g., due to filesystem specifics or shared layers),
		// we must stop to avoid an infinite loop on the same images.
		if vuAfter.UsedBytes >= vu.UsedBytes {
			break
		}

		// 9. Move to the next set of images.
		startIndex = n + 1
	}

	return report.Normalize(), nil
}

func (cleaner *LocalBackendCleaner) removeImageByRepoTags(ctx context.Context, options RunGCOptions, imgSummary image.Summary) (bool, error) {
	tagsCount := len(imgSummary.RepoTags)
	unRemovedCount := 0

	for _, ref := range imgSummary.RepoTags {
		// NOTE. <none:none> image is an intermediate image or a dandling image.
		// Here <none:none> image is the intermediate image.
		// We can assert this because we had remove all dandling images before via pruning.
		if ref == "<none>:<none>" {
			if err := cleaner.removeImageRef(ctx, ref, options); err != nil {
				logboek.Context(ctx).Info().LogF("Cannot remove local image by ID %q: %s\n", imgSummary.ID, err)
				unRemovedCount++
			}
		} else {
			if ok, err := cleaner.isLocked(container_backend.ImageLockName(ref)); err != nil {
				return false, fmt.Errorf("checking lock %q: %w", container_backend.ImageLockName(ref), err)
			} else if ok {
				unRemovedCount++
				continue
			}

			if err := cleaner.removeImageRef(ctx, ref, options); err != nil {
				logboek.Context(ctx).Info().LogF("Cannot remove local image by repo tag %q: %s\n", ref, err)
				unRemovedCount++
			}
		}
	}

	return unRemovedCount == 0 && tagsCount > 0, nil
}

func (cleaner *LocalBackendCleaner) removeImageByRepoDigests(ctx context.Context, options RunGCOptions, imgSummary image.Summary) (bool, error) {
	digestCount := len(imgSummary.RepoDigests)
	unRemovedCount := 0

	for _, repoDigest := range imgSummary.RepoDigests {
		if err := cleaner.removeImageRef(ctx, repoDigest, options); err != nil {
			logboek.Context(ctx).Info().LogF("Cannot remove local image by repo digest %q: %s\n", repoDigest, err)
			unRemovedCount++
		}
	}

	return unRemovedCount == 0 && digestCount > 0, nil
}

func (cleaner *LocalBackendCleaner) removeImageRef(ctx context.Context, ref string, options RunGCOptions) error {
	if options.DryRun {
		return nil
	}
	return cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{
		Force: options.Force,
	})
}

func (cleaner *LocalBackendCleaner) removeContainerRef(ctx context.Context, ref string, options RunGCOptions) error {
	if options.DryRun {
		return nil
	}
	return cleaner.backend.Rm(ctx, ref, container_backend.RmOpts{
		Force: options.Force,
	})
}

func (cleaner *LocalBackendCleaner) isLocked(lockName string) (bool, error) {
	ok, lock, err := cleaner.locker.Acquire(lockName, lockgate.AcquireOptions{NonBlocking: true})
	if err != nil {
		return false, err
	}
	if !ok {
		return true, nil
	}
	return false, cleaner.locker.Release(lock)
}

func logDeletedItems(ctx context.Context, deletedItems []string) {
	for _, item := range deletedItems {
		logboek.Context(ctx).LogLn(item)
	}
}

// handleError is common error handler
func handleError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, container_backend.ErrUnsupportedFeature):
		logboek.Context(ctx).Info().LogLn("Backend does not support this feature")
		return nil
	case errors.Is(err, errOptionDryRunNotSupported):
		logboek.Context(ctx).Info().LogLn("There is not able to calculate reclaimed size in --dry-run mode")
		return nil
	case err != nil:
		return err
	}
	return nil
}

// countImagesToFree returns index ∈ [0, len(images) - 1] or -1 otherwise.
// NOTE: img.Size contains the size of the image without considering shared layers.
// Thus, this counting relies on a lower-bound estimate of the space that will actually be freed.
// This is an intentional approximation used as an optimization to avoid repeatedly checking
// disk usage after deleting every single image.
func countImagesToFree(list image.ImagesList, startIndex int, bytesToFree uint64) int {
	if bytesToFree == 0 || startIndex < 0 || startIndex >= len(list) {
		return -1
	}

	i := startIndex

	for freedBytes := uint64(0); i < len(list) && freedBytes < bytesToFree; i++ {
		freedBytes += uint64(list[i].Size)
	}

	return int(math.Max(float64(i-1), float64(startIndex)))
}
