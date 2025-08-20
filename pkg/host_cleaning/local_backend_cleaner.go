package host_cleaning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/containers/image/v5/docker/reference"
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
	AllowedStorageVolumeUsagePercentage       float64
	AllowedStorageVolumeUsageMarginPercentage float64
	StoragePath                               string
	Force                                     bool
	DryRun                                    bool
}

type RunAutoGCOptions struct {
	AllowedStorageVolumeUsagePercentage float64
	StoragePath                         string
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

	switch backend.(type) {
	case *container_backend.DockerServerBackend:
		cleaner.backendType = containerBackendDocker
		return cleaner, nil
	case *container_backend.BuildahBackend:
		cleaner.backendType = containerBackendBuildah
		return cleaner, nil
	default:
		// returns cleaner for testing with mock
		cleaner.backendType = containerBackendTest
		return cleaner, ErrUnsupportedContainerBackend
	}
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

	// assert permissions
	if _, err := os.Stat(backendStoragePath); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
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
	return vu.Percentage() > options.AllowedStorageVolumeUsagePercentage, nil
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

	images := make(image.ImagesList, 0, len(list))

skipImage:
	for _, img := range list {
		projectName := img.Labels[image.WerfLabel]

		for _, ref := range img.RepoTags {
			normalizedTag, err := cleaner.normalizeReference(ref)
			if err != nil {
				return nil, err
			}

			// Do not remove stages-storage=:local images, because this is primary stages storage data,
			// and it can only be cleaned by the werf-cleanup command
			if strings.HasPrefix(normalizedTag, fmt.Sprintf("%s:", projectName)) {
				continue skipImage
			}
		}

		images = append(images, img)
	}

	return images, nil
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

	images := make(image.ImagesList, 0, len(list))

	// Do not remove stages-storage=:local images, because this is primary stages storage data,
	// and it can only be cleaned by the werf-cleanup command
skipImage:
	for _, img := range list {
		for _, ref := range img.RepoTags {
			normalizedTag, err := cleaner.normalizeReference(ref)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(normalizedTag, "werf-stages-storage/") {
				continue skipImage
			}
		}

		images = append(images, img)
	}

	return images, nil
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
			util.NewPair("reference", "*import-metadata-*"),
			util.NewPair("reference", "*-rejected"),

			util.NewPair("reference", "werf-client-id/*"),
			util.NewPair("reference", "werf-managed-images/*"),
			util.NewPair("reference", "werf-images-metadata-by-commit/*"),
			util.NewPair("reference", "werf-import-metadata/*"),
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

// normalizeReference Normalizes image reference (repository tag) to docker backend repository tag format.
func (cleaner *LocalBackendCleaner) normalizeReference(ref string) (string, error) {
	switch cleaner.backendType {
	case containerBackendDocker, containerBackendTest:
		return ref, nil
	case containerBackendBuildah:
		// ------------
		// WORKAROUND for Buildah
		// ------------
		// Buildah repository tag contains hostname (domain) prefix currently.
		// Example (buildah repo tag): localhost/werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265965865
		// We need normalize the tag to docker image repository tag format because our host cleanup algorithm based on docker backend.
		// Example: (docker repo tag): werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011
		// https://flant.kaiten.ru/space/193531/boards/card/26364854?focus=comment&focusId=56944076
		// TODO: unify repository tags in v3 on build stage

		named, err := reference.ParseNamed(ref)
		if err != nil {
			return "", err
		}
		hostnamePrefix := fmt.Sprintf("%s/", reference.Domain(named))

		return strings.TrimPrefix(ref, hostnamePrefix), nil
	default:
		return "", ErrUnsupportedContainerBackend
	}
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

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
	}

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).LogBlock("Check storage").Do(func() {
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage()), utils.BlueF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	targetVolumeUsagePercentage := math.Max(options.AllowedStorageVolumeUsagePercentage-options.AllowedStorageVolumeUsageMarginPercentage, 0)

	logboek.Context(ctx).LogBlock("Check storage").Do(func() {
		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsagePercentage))
		logboek.Context(ctx).LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(calcBytesToFree(vu, targetVolumeUsagePercentage))))
	})

	vuBefore := vu

	// Step 1. Prune unused anonymous volumes
	err = logboek.Context(ctx).LogBlock("Prune all unused anonymous volumes").DoError(func() error {
		reportVolumes, err := cleaner.pruneVolumes(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		vu.UsedBytes -= reportVolumes.SpaceReclaimed

		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportVolumes.SpaceReclaimed)))
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

		vu.UsedBytes -= reportImages.SpaceReclaimed

		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportImages.SpaceReclaimed)))
		logDeletedItems(ctx, reportImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune werf dangling images: %w", err)
	}

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).LogBlock("Check storage").Do(func() {
			logboek.Context(ctx).LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(vuBefore.UsedBytes-vu.UsedBytes)))
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsagePercentage))
		})

		return nil
	}

	// Step 3. Remove werf containers
	err = logboek.Context(ctx).LogBlock("Cleanup werf containers").DoError(func() error {
		reportWerfContainers, err := cleaner.safeCleanupWerfContainers(ctx, options, vu)
		if err != nil {
			return err
		}

		vu.UsedBytes -= reportWerfContainers.SpaceReclaimed

		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportWerfContainers.SpaceReclaimed)))
		logDeletedItems(ctx, reportWerfContainers.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to remove werf containers: %w", err)
	}

	// Step 4. Remove werf images
	err = logboek.Context(ctx).LogBlock("Cleanup werf images").DoError(func() error {
		reportWerfImages, err := cleaner.safeCleanupWerfImages(ctx, options, vu, targetVolumeUsagePercentage)
		if err != nil {
			return err
		}

		vu.UsedBytes -= reportWerfImages.SpaceReclaimed

		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportWerfImages.SpaceReclaimed)))
		logDeletedItems(ctx, reportWerfImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to cleanup werf images: %w", err)
	}

	logboek.Context(ctx).LogBlock("Check storage").Do(func() {
		logboek.Context(ctx).LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(vuBefore.UsedBytes-vu.UsedBytes)))
		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsagePercentage))
	})

	if vu.Percentage() > targetVolumeUsagePercentage {
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
			return newCleanupReport(), err
		}
		return mapImageListToCleanupReport(list), nil
	}

	report, err := cleaner.backend.PruneImages(ctx, prune.Options{Filters: filters})
	switch {
	case errors.Is(err, container_backend.ErrImageUsedByContainer),
		errors.Is(err, container_backend.ErrPruneIsAlreadyRunning):
		logboek.Context(ctx).Info().LogF("NOTE: Ignore image pruning: %s\n", err.Error())
		return newCleanupReport(), nil
	case err != nil:
		return newCleanupReport(), err
	}

	return mapPruneReportToCleanupReport(report), err
}

// pruneVolumes removes all anonymous volumes not used by at least one container
func (cleaner *LocalBackendCleaner) pruneVolumes(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		// NOTE: Buildah does not give us a way to precalculate pruned size.
		// NOTE: Docker does not give us a way to precalculate pruned size.
		return newCleanupReport(), errOptionDryRunNotSupported
	}

	report, err := cleaner.backend.PruneVolumes(ctx, prune.Options{})

	switch {
	case errors.Is(err, container_backend.ErrPruneIsAlreadyRunning):
		logboek.Context(ctx).Info().LogF("NOTE: Ignore volume pruning: %s\n", err.Error())
		return newCleanupReport(), nil
	case err != nil:
		return newCleanupReport(), err
	}

	return mapPruneReportToCleanupReport(report), err
}

// safeCleanupWerfContainers cleanups werf containers safely using host locks
func (cleaner *LocalBackendCleaner) safeCleanupWerfContainers(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage) (cleanupReport, error) {
	containers, err := werfContainersByContainersOptions(ctx, cleaner.backend, buildContainersOptions())
	if err != nil {
		return newCleanupReport(), fmt.Errorf("cannot get build containers: %w", err)
	}

	if options.DryRun {
		return mapContainerListToCleanupReport(containers), nil
	}

	report, err := cleaner.doSafeCleanupWerfContainers(ctx, options, vu, containers)
	if err != nil {
		return newCleanupReport(), err
	}

	return report, nil
}

func (cleaner *LocalBackendCleaner) doSafeCleanupWerfContainers(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage, containers image.ContainerList) (cleanupReport, error) {
	report := newCleanupReport()
	report.ItemsDeleted = make([]string, 0, len(containers))

	for _, container := range containers {
		containerName := werfContainerName(container)

		if containerName == "" {
			logboek.Context(ctx).Info().LogF("Ignore bad container %s\n", container.ID)
			continue
		}

		if ok, err := cleaner.isLocked(container_backend.ContainerLockName(containerName)); err != nil {
			return newCleanupReport(), fmt.Errorf("checking lock %q: %w", container_backend.ContainerLockName(containerName), err)
		} else if ok {
			continue
		}

		if err := cleaner.backend.Rm(ctx, container.ID, container_backend.RmOpts{Force: options.Force}); err != nil {
			switch {
			case errors.Is(err, container_backend.ErrCannotRemovePausedContainer):
				logboek.Context(ctx).Info().LogF("Ignore paused container %s\n", logContainerName(container))
				return newCleanupReport(), nil
			case errors.Is(err, container_backend.ErrCannotRemoveRunningContainer):
				logboek.Context(ctx).Info().LogF("Ignore running container %s\n", logContainerName(container))
				return newCleanupReport(), nil
			case err != nil:
				return newCleanupReport(), fmt.Errorf("failed to remove container %s: %w", logContainerName(container), err)
			}
		}

		report.ItemsDeleted = append(report.ItemsDeleted, container.ID)
	}

	if len(containers) > 0 {
		// No explicit way to calculate reclaimed containers' size.
		// But we can calculate it implicitly via disk usage check.
		vuAfter, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return cleanupReport{}, fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}
		report.SpaceReclaimed = vu.UsedBytes - vuAfter.UsedBytes
	}

	return report, nil
}

// safeCleanupWerfImages cleanups werf images safely using host locks
func (cleaner *LocalBackendCleaner) safeCleanupWerfImages(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64) (cleanupReport, error) {
	images, err := cleaner.werfImages(ctx)
	if err != nil {
		return newCleanupReport(), err
	}

	if options.DryRun {
		n := countImagesToFree(images, 0, calcBytesToFree(vu, targetVolumeUsagePercentage))
		if n == -1 {
			return newCleanupReport(), nil
		}
		return mapImageListToCleanupReport(images[:n+1]), nil
	}

	report, err := cleaner.doSafeCleanupWerfImages(ctx, options, vu, targetVolumeUsagePercentage, images)
	if err != nil {
		return newCleanupReport(), err
	}

	return report, nil
}

func (cleaner *LocalBackendCleaner) doSafeCleanupWerfImages(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64, images image.ImagesList) (cleanupReport, error) {
	report := newCleanupReport()
	report.ItemsDeleted = make([]string, 0, len(images))

	tVu := targetVolumeUsagePercentage

	for i, n := 0, countImagesToFree(images, 0, calcBytesToFree(vu, tVu)); n != -1; i, n = n+1, countImagesToFree(images, n+1, calcBytesToFree(vu, tVu)) {

		for _, imgSummary := range images[i : n+1] {

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

		// Here we actually don't know how much space was reclaimed. There are several reasons of why:
		//	1. Image can have "shared size" but backend cannot calculate the shared size properly.
		//  2. An error can happen while removing of img.RepoTags of img.RepoDigests.
		//
		// So there is only one way to handle this is re-calculate of disk usage size (expensive operation):
		vuAfter, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return report, fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}

		report.SpaceReclaimed += vu.UsedBytes - vuAfter.UsedBytes
		// we must update vu variable to re-calculate how many images need to clean-up
		vu = vuAfter
	}

	return report, nil
}

func (cleaner *LocalBackendCleaner) removeImageByRepoTags(ctx context.Context, options RunGCOptions, imgSummary image.Summary) (bool, error) {
	tagsCount := len(imgSummary.RepoTags)
	unRemovedCount := 0

	for _, ref := range imgSummary.RepoTags {
		// NOTE. <none:none> image is an intermediate image or a dandling image.
		// Here <none:none> image is the intermediate image.
		// We can assert this because we had remove all dandling images before via pruning.
		if ref == "<none>:<none>" {
			err := cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{
				Force: options.Force,
			})
			if err != nil {
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

			if err := cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{Force: options.Force}); err != nil {
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
		err := cleaner.backend.Rmi(ctx, repoDigest, container_backend.RmiOpts{
			Force: options.Force,
		})
		if err != nil {
			logboek.Context(ctx).Info().LogF("Cannot remove local image by repo digest %q: %s\n", repoDigest, err)
			unRemovedCount++
		}
	}

	return unRemovedCount == 0 && digestCount > 0, nil
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

func calcBytesToFree(vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64) uint64 {
	diffPercentage := vu.Percentage() - targetVolumeUsagePercentage
	allowedVolumeUsageToFree := math.Max(diffPercentage, 0)
	bytesToFree := uint64((float64(vu.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
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

// countImagesToFree returns index ∈ [0, len(images) - 1] or -1 otherwise
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
