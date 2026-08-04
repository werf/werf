package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/slug"
)

const (
	RepoMetaRepoMarker_ImageTagPrefix  = "werf-meta-repo-marker-"
	RepoMetaRepoMarker_ImageNameFormat = "%s:werf-meta-repo-marker-%s"
	RepoMetaRepoMarker_LabelMetaRepo   = "metarepoaddress"
)

// CanonicalRepoAddress normalizes a registry repository address so equivalent
// spellings (docker.io/x vs index.docker.io/library/x) compare equal.
func CanonicalRepoAddress(address string) (string, error) {
	addr, err := ref.ParseAddr(address)
	if err != nil {
		return "", fmt.Errorf("parse repo address %q: %w", address, err)
	}
	if addr.RegistryAddress == nil {
		return "", fmt.Errorf("%q is not a container registry address", address)
	}
	return addr.RegistryAddress.Repo, nil
}

func getMetaRepoMarkerID(projectName string) string {
	id := slugImageName(projectName)
	if !slug.IsValidDockerTag(id) {
		id = slug.LimitedSlug(id, slug.DockerTagMaxSize-len(RepoMetaRepoMarker_ImageTagPrefix))
	}
	return id
}

func makeRepoMetaRepoMarkerRecord(repoAddress, projectName string) string {
	return fmt.Sprintf(RepoMetaRepoMarker_ImageNameFormat, repoAddress, getMetaRepoMarkerID(projectName))
}

// GetMetaRepoMarker returns the canonical meta-repo address recorded for the
// project in this repo. found is false when no marker exists; found is true with
// an empty address when the marker is present but malformed.
func (storage *RepoStagesStorage) GetMetaRepoMarker(ctx context.Context, projectName string) (address string, found bool, err error) {
	fullImageName := makeRepoMetaRepoMarkerRecord(storage.RepoAddress, projectName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return "", false, fmt.Errorf("unable to get meta-repo marker %q: %w", fullImageName, err)
	}
	if img == nil {
		return "", false, nil
	}
	return img.Labels[RepoMetaRepoMarker_LabelMetaRepo], true, nil
}

func (storage *RepoStagesStorage) PutMetaRepoMarker(ctx context.Context, projectName, metaRepoAddress string) error {
	fullImageName := makeRepoMetaRepoMarkerRecord(storage.RepoAddress, projectName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{
		image.WerfLabel:                  projectName,
		RepoMetaRepoMarker_LabelMetaRepo: metaRepoAddress,
	}}
	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push meta-repo marker %q: %w", fullImageName, err)
	}
	return nil
}

func (storage *RepoStagesStorage) RmMetaRepoMarker(ctx context.Context, projectName string) error {
	fullImageName := makeRepoMetaRepoMarkerRecord(storage.RepoAddress, projectName)

	imgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get meta-repo marker %q: %w", fullImageName, err)
	}
	if imgInfo == nil {
		return nil
	}
	if err := storage.DockerRegistry.DeleteRepoImage(ctx, imgInfo); err != nil {
		return fmt.Errorf("unable to delete meta-repo marker %q: %w", fullImageName, err)
	}
	return nil
}

func isMetadataCandidateTag(tag string) bool {
	if strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
		return false
	}
	switch {
	case strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix):
		return true
	case strings.HasPrefix(tag, RepoCustomTagMetadata_ImageTagPrefix):
		return true
	case tag == RepoCleanUpRecord_ImageTagPrefix:
		return true
	case strings.HasPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix):
		return len(strings.Split(strings.TrimPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix), "_")) == 3
	default:
		return false
	}
}

type projectMetadataRecord struct {
	tag  string
	info *image.Info
}

// metadataRecordMatchesProject decides whether a metadata candidate record
// belongs to the project. The cleanup record was introduced already labeled and
// timestamped, so it is matched strictly. managed-image, image-metadata and
// custom-tag-metadata records were historically pushed without the werf label
// (and the read paths that consult them are label-blind), so ownership is
// ambiguous: an unlabeled record is conservatively treated as the project's (a
// foreign label excludes it), matching the plan's requirement to refuse adoption
// rather than orphan metadata that cleanup still reads.
func metadataRecordMatchesProject(tag string, labels map[string]string, projectName string) bool {
	owner, hasOwner := labels[image.WerfLabel]
	if tag == RepoCleanUpRecord_ImageTagPrefix {
		_, hasTimestamp := labels[RepoCleanUpRecord_LabelTimestamp]
		return hasTimestamp && owner == projectName
	}
	return !hasOwner || owner == "" || owner == projectName
}

// collectProjectMetadataRecords returns the metadata records in this repo owned
// by the project (the four families routed through the meta-repo), per
// metadataRecordMatchesProject. When stopOnFirst is set it returns after the
// first match (cheap existence probe).
func (storage *RepoStagesStorage) collectProjectMetadataRecords(ctx context.Context, projectName string, stopOnFirst bool) ([]projectMetadataRecord, error) {
	tags, err := storage.Tags(ctx, storage.RepoAddress, docker_registry.WithCachedTags())
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	var res []projectMetadataRecord
	for _, tag := range tags {
		if !isMetadataCandidateTag(tag) {
			continue
		}

		fullImageName := fmt.Sprintf("%s:%s", storage.RepoAddress, tag)
		img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
		if err != nil {
			return nil, fmt.Errorf("unable to get repo image %q: %w", fullImageName, err)
		}
		if img == nil || !metadataRecordMatchesProject(tag, img.Labels, projectName) {
			continue
		}

		res = append(res, projectMetadataRecord{tag: tag, info: img})
		if stopOnFirst {
			return res, nil
		}
	}
	return res, nil
}

func (storage *RepoStagesStorage) HasProjectMetadata(ctx context.Context, projectName string) (bool, error) {
	records, err := storage.collectProjectMetadataRecords(ctx, projectName, true)
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

func (storage *RepoStagesStorage) ensureMetaRepoMarker(ctx context.Context, projectName, metaRepoAddress string) error {
	existing, found, err := storage.GetMetaRepoMarker(ctx, projectName)
	if err != nil {
		return err
	}
	if found {
		if existing == metaRepoAddress {
			return nil
		}
		return fmt.Errorf("meta-repo marker for project %q in %s points at %q, but %q is being used — they must match", projectName, storage.RepoAddress, existing, metaRepoAddress)
	}
	return storage.PutMetaRepoMarker(ctx, projectName, metaRepoAddress)
}

// metaRepoMarkerStorage decorates a distinct meta storage so the first real
// metadata write plants the per-project safeguard marker into the stages repo.
type metaRepoMarkerStorage struct {
	PrimaryStagesStorage

	markerStore     *RepoStagesStorage
	projectName     string
	metaRepoAddress string
	cleanupDisabled bool

	once    sync.Once
	onceErr error
}

var _ PrimaryStagesStorage = (*metaRepoMarkerStorage)(nil)

func (s *metaRepoMarkerStorage) ensure(ctx context.Context) error {
	s.once.Do(func() {
		s.onceErr = s.markerStore.ensureMetaRepoMarker(ctx, s.projectName, s.metaRepoAddress)
	})
	return s.onceErr
}

func (s *metaRepoMarkerStorage) AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.AddManagedImage(ctx, projectName, imageNameOrManagedImageName)
}

func (s *metaRepoMarkerStorage) RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.RmManagedImage(ctx, projectName, imageNameOrManagedImageName)
}

func (s *metaRepoMarkerStorage) PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.PutImageMetadata(ctx, projectName, imageNameOrManagedImageName, commit, stageID)
}

func (s *metaRepoMarkerStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.RmImageMetadata(ctx, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID)
}

func (s *metaRepoMarkerStorage) RegisterStageCustomTag(ctx context.Context, projectName string, stageDesc *image.StageDesc, tag string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.RegisterStageCustomTag(ctx, projectName, stageDesc, tag)
}

func (s *metaRepoMarkerStorage) UnregisterStageCustomTag(ctx context.Context, tag string) error {
	if err := s.ensure(ctx); err != nil {
		return err
	}
	return s.PrimaryStagesStorage.UnregisterStageCustomTag(ctx, tag)
}

func (s *metaRepoMarkerStorage) PostLastCleanupRecord(ctx context.Context, projectName string) error {
	if !s.cleanupDisabled {
		if err := s.ensure(ctx); err != nil {
			return err
		}
	}
	return s.PrimaryStagesStorage.PostLastCleanupRecord(ctx, projectName)
}

const (
	metaRepoMismatchMsg    = "metadata for project %q is stored in meta-repo %q (per the marker in %s), but --meta-repo=%q was given — they must match"
	metaRepoMissingFlagMsg = "metadata for project %q is stored in a separate meta-repo %q (per the marker in %s); pass --meta-repo %q, or run 'werf meta-repo detach' to remove the safeguard"
	metaRepoMigrateMsg     = "--repo %q already contains metadata for project %q; run 'werf meta-repo migrate --repo %q --meta-repo %q' to move it before using --meta-repo"
	metaRepoMalformedMsg   = "the meta-repo marker for project %q in %s is malformed (no meta-repo address); run 'werf meta-repo detach' to remove it"
)

// SetupMetaRepoSafeguard validates the per-project meta-repo marker and, when a
// distinct meta-repo is in use, returns a decorated meta storage that plants the
// marker on the first real metadata write. It is a no-op for non-registry stages
// storage and when --meta-repo canonicalizes to --repo.
func SetupMetaRepoSafeguard(ctx context.Context, projectName string, stagesStorage, metaStorage PrimaryStagesStorage, cleanupDisabled bool) (PrimaryStagesStorage, error) {
	repoStages, ok := stagesStorage.(*RepoStagesStorage)
	if !ok {
		return metaStorage, nil
	}

	noFlag := metaStorage == stagesStorage

	var metaConfigured bool
	var metaCanonical string
	if !noFlag {
		repoCanonical, err := CanonicalRepoAddress(repoStages.Address())
		if err != nil {
			return nil, err
		}
		metaCanonical, err = CanonicalRepoAddress(metaStorage.Address())
		if err != nil {
			return nil, err
		}
		metaConfigured = metaCanonical != repoCanonical
	}

	if !noFlag && !metaConfigured {
		return metaStorage, nil
	}

	markerAddr, markerFound, err := repoStages.GetMetaRepoMarker(ctx, projectName)
	if err != nil {
		return nil, err
	}

	if markerFound && markerAddr == "" {
		return nil, fmt.Errorf(metaRepoMalformedMsg, projectName, repoStages.Address())
	}

	if !metaConfigured {
		if markerFound {
			return nil, fmt.Errorf(metaRepoMissingFlagMsg, projectName, markerAddr, repoStages.Address(), markerAddr)
		}
		return metaStorage, nil
	}

	if markerFound {
		if markerAddr != metaCanonical {
			return nil, fmt.Errorf(metaRepoMismatchMsg, projectName, markerAddr, repoStages.Address(), metaStorage.Address())
		}
	} else {
		has, err := repoStages.HasProjectMetadata(ctx, projectName)
		if err != nil {
			return nil, err
		}
		if has {
			return nil, fmt.Errorf(metaRepoMigrateMsg, repoStages.Address(), projectName, repoStages.Address(), metaStorage.Address())
		}
	}

	return &metaRepoMarkerStorage{
		PrimaryStagesStorage: metaStorage,
		markerStore:          repoStages,
		projectName:          projectName,
		metaRepoAddress:      metaCanonical,
		cleanupDisabled:      cleanupDisabled,
	}, nil
}

type MigrateMetaRepoOptions struct {
	RemoveSource bool
}

// MigrateMetaRepo copies the project's metadata records from the source stages
// repo into the destination meta-repo (copy-first, idempotent), plants the
// marker in the source, and — with RemoveSource — deletes each source original
// only after its destination copy is verified present.
func MigrateMetaRepo(ctx context.Context, projectName string, src, dst *RepoStagesStorage, opts MigrateMetaRepoOptions) error {
	srcCanonical, err := CanonicalRepoAddress(src.Address())
	if err != nil {
		return err
	}
	dstCanonical, err := CanonicalRepoAddress(dst.Address())
	if err != nil {
		return err
	}
	if srcCanonical == dstCanonical {
		return fmt.Errorf("--meta-repo %q resolves to the same repository as --repo %q; nothing to migrate", dst.Address(), src.Address())
	}

	markerAddr, markerFound, err := src.GetMetaRepoMarker(ctx, projectName)
	if err != nil {
		return err
	}
	if markerFound && markerAddr != "" && markerAddr != dstCanonical {
		return fmt.Errorf("--repo %q already has a meta-repo safeguard for project %q pointing at %q; refusing to migrate to %q", src.Address(), projectName, markerAddr, dstCanonical)
	}

	records, err := src.collectProjectMetadataRecords(ctx, projectName, false)
	if err != nil {
		return err
	}

	for _, rec := range records {
		dstRef := fmt.Sprintf("%s:%s", dst.Address(), rec.tag)
		existing, err := dst.DockerRegistry.TryGetRepoImage(ctx, dstRef)
		if err != nil {
			return fmt.Errorf("unable to check destination %q: %w", dstRef, err)
		}
		if existing != nil {
			if !metadataRecordMatchesProject(rec.tag, existing.Labels, projectName) {
				return fmt.Errorf("destination %q already exists but is not a valid metadata record owned by project %q; refusing to migrate", dstRef, projectName)
			}
			logboek.Context(ctx).Info().LogF("Skipping %s (already present in meta-repo)\n", rec.tag)
			continue
		}

		srcRef := fmt.Sprintf("%s:%s", src.Address(), rec.tag)
		if err := dst.DockerRegistry.CopyImage(ctx, srcRef, dstRef, docker_registry.CopyImageOptions{}); err != nil {
			return fmt.Errorf("unable to copy %s to %s: %w", srcRef, dstRef, err)
		}
		logboek.Context(ctx).Info().LogF("Copied %s to meta-repo\n", rec.tag)
	}

	if err := src.PutMetaRepoMarker(ctx, projectName, dstCanonical); err != nil {
		return err
	}

	if !opts.RemoveSource {
		return nil
	}

	for _, rec := range records {
		dstRef := fmt.Sprintf("%s:%s", dst.Address(), rec.tag)
		img, err := dst.DockerRegistry.TryGetRepoImage(ctx, dstRef)
		if err != nil {
			return fmt.Errorf("unable to verify destination %q: %w", dstRef, err)
		}
		if img == nil || !metadataRecordMatchesProject(rec.tag, img.Labels, projectName) {
			return fmt.Errorf("destination %q missing or not a valid record for project %q after copy; not deleting source", dstRef, projectName)
		}
		if err := src.DockerRegistry.DeleteRepoImage(ctx, rec.info); err != nil {
			return fmt.Errorf("unable to delete source %q: %w", rec.info.Name, err)
		}
		logboek.Context(ctx).Info().LogF("Removed source %s\n", rec.tag)
	}
	return nil
}
