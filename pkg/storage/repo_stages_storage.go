package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/util"
)

const (
	RepoStage_ImageFormatWithUniqueID = "%s:%s-%d"
	RepoStage_ImageFormat             = "%s:%s"

	RepoManagedImageRecord_ImageTagPrefix  = "managed-image-"
	RepoManagedImageRecord_ImageNameFormat = "%s:managed-image-%s"

	RepoRejectedStageImageRecord_ImageTagSuffix  = "-rejected"
	RepoRejectedStageImageRecord_ImageNameFormat = "%s:%s-%d-rejected"

	RepoImageMetadataByCommitRecord_ImageTagPrefix = "meta-"
	RepoImageMetadataByCommitRecord_TagFormat      = "meta-%s_%s_%s"

	RepoCustomTagMetadata_ImageTagPrefix  = "custom-tag-meta-"
	RepoCustomTagMetadata_ImageNameFormat = "%s:custom-tag-meta-%s"

	RepoImportMetadata_ImageTagPrefix  = "import-metadata-"
	RepoImportMetadata_ImageNameFormat = "%s:import-metadata-%s"

	RepoClientIDRecord_ImageTagPrefix  = "client-id-"
	RepoClientIDRecord_ImageNameFormat = "%s:client-id-%s-%d"

	UnexpectedTagFormatErrorPrefix = "unexpected tag format"
)

func getDigestAndUniqueIDFromRepoStageImageTag(repoStageImageTag string) (string, int64, error) {
	parts := strings.SplitN(repoStageImageTag, "-", 2)

	if len(parts) == 1 {
		if len(parts[0]) != 56 {
			return "", 0, fmt.Errorf("%s: sha3-224 hash expected, got %q", UnexpectedTagFormatErrorPrefix, parts[0])
		}
		return parts[0], 0, nil
	} else if len(parts) != 2 {
		return "", 0, fmt.Errorf("%s %s", UnexpectedTagFormatErrorPrefix, repoStageImageTag)
	}

	if uniqueID, err := image.ParseUniqueIDAsTimestamp(parts[1]); err != nil {
		return "", 0, fmt.Errorf("%s %s: unable to parse unique id %s as timestamp: %w", UnexpectedTagFormatErrorPrefix, repoStageImageTag, parts[1], err)
	} else {
		return parts[0], uniqueID, nil
	}
}

func isUnexpectedTagFormatError(err error) bool {
	return strings.HasPrefix(err.Error(), UnexpectedTagFormatErrorPrefix)
}

type RepoStagesStorage struct {
	RepoAddress      string
	DockerRegistry   docker_registry.Interface
	ContainerBackend container_backend.ContainerBackend
}

func NewRepoStagesStorage(repoAddress string, containerBackend container_backend.ContainerBackend, dockerRegistry docker_registry.Interface) *RepoStagesStorage {
	return &RepoStagesStorage{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerBackend: containerBackend,
	}
}

func (storage *RepoStagesStorage) ConstructStageImageName(_, digest string, uniqueID int64) string {
	if uniqueID == 0 {
		return fmt.Sprintf(RepoStage_ImageFormat, storage.RepoAddress, digest)
	}
	return fmt.Sprintf(RepoStage_ImageFormatWithUniqueID, storage.RepoAddress, digest, uniqueID)
}

func (storage *RepoStagesStorage) GetStagesIDs(ctx context.Context, _ string, opts ...Option) ([]image.StageID, error) {
	var res []image.StageID

	o := makeOptions(opts...)
	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %w", storage.RepoAddress, err)
	} else {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetStagesIDs fetched tags for %q: %#v\n", storage.RepoAddress, tags)

		for _, tag := range tags {
			isRegularStage := (len(tag) == 70 && len(strings.Split(tag, "-")) == 2) // 2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7-1611836746968
			isMultiplatformStage := (len(tag) == 56)                                // 2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7
			if !isRegularStage && !isMultiplatformStage {
				continue
			}

			if strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix) || strings.HasPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix) || strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			if digest, uniqueID, err := getDigestAndUniqueIDFromRepoStageImageTag(tag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Debug().LogLn(err.Error())
					continue
				}
				return nil, err
			} else {
				res = append(res, *image.NewStageID(digest, uniqueID))

				logboek.Context(ctx).Debug().LogF("Selected stage by digest %q uniqueID %d\n", digest, uniqueID)
			}
		}

		return res, nil
	}
}

func (storage *RepoStagesStorage) ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error {
	return storage.DockerRegistry.MutateAndPushImage(ctx, stageDescription.Info.Name, destinationReference, mutateExportStageConfig(mutateConfigFunc))
}

func mutateExportStageConfig(mutateConfigFunc func(config v1.Config) (v1.Config, error)) func(config v1.Config) (v1.Config, error) {
	return func(config v1.Config) (v1.Config, error) {
		if config.Labels == nil {
			panic("unexpected condition: stage image without labels")
		}

		for name := range config.Labels {
			if strings.HasPrefix(name, image.WerfLabelPrefix) {
				delete(config.Labels, name)
			}
		}

		if mutateConfigFunc != nil {
			return mutateConfigFunc(config)
		}

		return config, nil
	}
}

func (storage *RepoStagesStorage) DeleteStage(ctx context.Context, stageDescription *image.StageDescription, _ DeleteImageOptions) error {
	if err := storage.DockerRegistry.DeleteRepoImage(ctx, stageDescription.Info); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %w", stageDescription.Info.Name, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, stageDescription.StageID.Digest, stageDescription.StageID.UniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.DeleteStage full image name: %s\n", rejectedImageName)

	if rejectedImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, rejectedImageName); err != nil {
		return fmt.Errorf("unable to get rejected image record %q: %w", rejectedImageName, err)
	} else if rejectedImgInfo != nil {
		if err := storage.DockerRegistry.DeleteRepoImage(ctx, rejectedImgInfo); err != nil {
			return fmt.Errorf("unable to remove rejected image record %q: %w", rejectedImageName, err)
		}
	}

	return nil
}

func makeRepoRejectedStageImageRecord(repoAddress, digest string, uniqueID int64) string {
	return fmt.Sprintf(RepoRejectedStageImageRecord_ImageNameFormat, repoAddress, digest, uniqueID)
}

func (storage *RepoStagesStorage) RejectStage(ctx context.Context, projectName, digest string, uniqueID int64) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RejectStage %s %s %d\n", projectName, digest, uniqueID)

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, digest, uniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RejectStage full image name: %s\n", rejectedImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}
	if err := storage.DockerRegistry.PushImage(ctx, rejectedImageName, opts); err != nil {
		return fmt.Errorf("unable to push rejected stage image record %s: %w", rejectedImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Rejected stage by digest %s uniqueID %d\n", digest, uniqueID)
	return nil
}

func (storage *RepoStagesStorage) CreateRepo(ctx context.Context) error {
	return storage.DockerRegistry.CreateRepo(ctx, storage.RepoAddress)
}

func (storage *RepoStagesStorage) DeleteRepo(ctx context.Context) error {
	return storage.DockerRegistry.DeleteRepo(ctx, storage.RepoAddress)
}

func (storage *RepoStagesStorage) GetStagesIDsByDigest(ctx context.Context, _, digest string, opts ...Option) ([]image.StageID, error) {
	var res []image.StageID

	o := makeOptions(opts...)
	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %w", storage.RepoAddress, err)
	} else {
		var rejectedStages []image.StageID

		for _, tag := range tags {
			if !strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			realTag := strings.TrimSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix)

			if _, uniqueID, err := getDigestAndUniqueIDFromRepoStageImageTag(realTag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Info().LogF("Unexpected tag %q format: %s\n", realTag, err)
					continue
				}
				return nil, fmt.Errorf("unable to get digest and uniqueID from rejected stage tag %q: %w", tag, err)
			} else {
				logboek.Context(ctx).Info().LogF("Found rejected stage %q\n", tag)
				rejectedStages = append(rejectedStages, *image.NewStageID(digest, uniqueID))
			}
		}

	FindSuitableStages:
		for _, tag := range tags {
			if !strings.HasPrefix(tag, digest) {
				continue
			}

			if strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			if _, uniqueID, err := getDigestAndUniqueIDFromRepoStageImageTag(tag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Debug().LogLn(err.Error())
					logboek.Context(ctx).Info().LogF("Unexpected tag %q format: %s\n", tag, err)
					continue
				}
				return nil, fmt.Errorf("unable to get digest and uniqueID from tag %q: %w", tag, err)
			} else {
				stageID := image.NewStageID(digest, uniqueID)

				for _, rejectedStage := range rejectedStages {
					if rejectedStage.Digest == stageID.Digest && rejectedStage.UniqueID == stageID.UniqueID {
						logboek.Context(ctx).Info().LogF("Discarding rejected stage %q\n", tag)
						continue FindSuitableStages
					}
				}

				logboek.Context(ctx).Debug().LogF("Stage %q is suitable for digest %q\n", tag, digest)
				res = append(res, *stageID)
			}
		}
	}

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesByDigest result for %q: %#v\n", storage.RepoAddress, res)

	return res, nil
}

// NOTE: Do we need a special option for multiplatform image description getter?
// NOTE: go-containerregistry has a special method to get an v1.ImageIndex manifest, instead of v1.Image,
// NOTE:   should we utilize this method in GetStageDescription also?
// NOTE: Current version always tries to get v1.Image manifest and it works without errors though.
func (storage *RepoStagesStorage) GetStageDescription(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage GetStageDescription %s %s %d\n", projectName, stageID.Digest, stageID.UniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage stageImageName = %q\n", stageImageName)

	imgInfo, err := storage.DockerRegistry.GetRepoImage(ctx, stageImageName)
	if docker_registry.IsImageNotFoundError(err) {
		return nil, nil
	}
	if docker_registry.IsBrokenImageError(err) {
		return nil, ErrBrokenImage
	}
	if err != nil {
		return nil, fmt.Errorf("unable to inspect repo image %s: %w", stageImageName, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, stageID.Digest, stageID.UniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetStageDescription check rejected image name: %s\n", rejectedImageName)

	if rejectedImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, rejectedImageName); err != nil {
		return nil, fmt.Errorf("unable to get repo image %q: %w", rejectedImageName, err)
	} else if rejectedImgInfo != nil {
		logboek.Context(ctx).Info().LogF("Stage digest %s uniqueID %d image is rejected: ignore stage image\n", stageID.Digest, stageID.UniqueID)
		return nil, nil
	}

	return &image.StageDescription{
		StageID: image.NewStageID(stageID.Digest, stageID.UniqueID),
		Info:    imgInfo,
	}, nil
}

func (storage *RepoStagesStorage) CheckStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error {
	fullImageName := strings.Join([]string{storage.RepoAddress, tag}, ":")
	customTagImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return err
	}

	if customTagImgInfo == nil {
		return fmt.Errorf("custom tag %q not found", tag)
	}

	if customTagImgInfo.ID != stageDescription.Info.ID {
		return fmt.Errorf("custom tag %q image must be the same as associated content-based tag %q image", tag, stageDescription.StageID.String())
	}

	return nil
}

func (storage *RepoStagesStorage) AddStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error {
	return storage.DockerRegistry.TagRepoImage(ctx, stageDescription.Info, tag)
}

func (storage *RepoStagesStorage) DeleteStageCustomTag(ctx context.Context, tag string) error {
	fullImageName := strings.Join([]string{storage.RepoAddress, tag}, ":")
	imgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get repo image %q info: %w", fullImageName, err)
	}

	if imgInfo == nil {
		return nil
	}

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, imgInfo); err != nil {
		return fmt.Errorf("unable to delete image %q from repo: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) addStageCustomTagMetadata(ctx context.Context, projectName string, stageDescription *image.StageDescription, tag string) error {
	fullImageName := makeRepoCustomTagMetadataRecord(storage.RepoAddress, tag)
	metadata := newCustomTagMetadata(stageDescription.StageID.String(), tag)
	opts := &docker_registry.PushImageOptions{
		Labels: metadata.ToLabels(),
	}
	opts.Labels[image.WerfLabel] = projectName

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) deleteStageCustomTagMetadata(ctx context.Context, tagOrID string) error {
	fullImageName := makeRepoCustomTagMetadataRecord(storage.RepoAddress, tagOrID)
	imgInfo, err := storage.DockerRegistry.GetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get repo image %q info: %w", fullImageName, err)
	}

	if imgInfo == nil {
		panic("unexpected condition")
	}

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, imgInfo); err != nil {
		return fmt.Errorf("unable to delete image %q from repo: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) GetStageCustomTagMetadata(ctx context.Context, tagOrID string) (*CustomTagMetadata, error) {
	fullImageName := makeRepoCustomTagMetadataRecord(storage.RepoAddress, tagOrID)
	img, err := storage.DockerRegistry.GetRepoImage(ctx, fullImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo image %s: %w", fullImageName, err)
	}

	if img == nil {
		panic("unexpected condition")
	}

	return newCustomTagMetadataFromLabels(img.Labels), nil
}

func (storage *RepoStagesStorage) GetStageCustomTagMetadataIDs(ctx context.Context, opts ...Option) ([]string, error) {
	o := makeOptions(opts...)
	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	var res []string
	for _, tag := range tags {
		if !strings.HasPrefix(tag, RepoCustomTagMetadata_ImageTagPrefix) {
			continue
		}

		id := strings.TrimPrefix(tag, RepoCustomTagMetadata_ImageTagPrefix)
		res = append(res, id)
	}

	return res, nil
}

func (storage *RepoStagesStorage) RegisterStageCustomTag(ctx context.Context, projectName string, stageDescription *image.StageDescription, tag string) error {
	if err := storage.addStageCustomTagMetadata(ctx, projectName, stageDescription, tag); err != nil {
		return fmt.Errorf("unable to add stage custom tag metadata: %w", err)
	}
	return nil
}

func (storage *RepoStagesStorage) UnregisterStageCustomTag(ctx context.Context, tag string) error {
	if err := storage.deleteStageCustomTagMetadata(ctx, tag); err != nil {
		return fmt.Errorf("unable to delete stage custom tag metadata: %w", err)
	}
	return nil
}

func (storage *RepoStagesStorage) AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage %s %s\n", projectName, imageNameOrManagedImageName)

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageNameOrManagedImageName)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}
	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmManagedImage %s %s\n", projectName, imageNameOrManagedImageName)

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageNameOrManagedImageName)

	imgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get repo image %q info: %w", fullImageName, err)
	}

	if imgInfo == nil {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmManagedImage record %q does not exist => exiting\n", fullImageName)
		return nil
	}

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, imgInfo); err != nil {
		return fmt.Errorf("unable to delete image %q from repo: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) IsManagedImageExist(ctx context.Context, _, imageNameOrManagedImageName string, opts ...Option) (bool, error) {
	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageNameOrManagedImageName)
	o := makeOptions(opts...)
	return storage.DockerRegistry.IsTagExist(ctx, fullImageName, o.dockerRegistryOptions...)
}

func (storage *RepoStagesStorage) GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetManagedImages %s\n", projectName)

	o := makeOptions(opts...)
	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	var res []string
	for _, tag := range tags {
		if !strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix) {
			continue
		}

		managedImageName := getManagedImageNameFromManagedImageID(strings.TrimPrefix(tag, RepoManagedImageRecord_ImageTagPrefix))
		res = append(res, managedImageName)
	}

	return res, nil
}

func (storage *RepoStagesStorage) FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	if err := storage.ContainerBackend.PullImageFromRegistry(ctx, img); err != nil {
		if strings.HasSuffix(err.Error(), "unknown blob") {
			return ErrBrokenImage
		}
		return err
	}

	return nil
}

// FIXME(stapel-to-buildah): use ImageInterface instead of LegacyImageInterface
// FIXME(stapel-to-buildah): possible optimization would be to push buildah container directly into registry wihtout committing a local image
func (storage *RepoStagesStorage) StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	if img.BuiltID() != "" {
		if err := storage.ContainerBackend.Tag(ctx, img.BuiltID(), img.Name(), container_backend.TagOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
			return fmt.Errorf("unable to tag built image %q by %q: %w", img.BuiltID(), img.Name(), err)
		}
	}

	if err := storage.ContainerBackend.Push(ctx, img.Name(), container_backend.PushOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return fmt.Errorf("unable to push image %q: %w", img.Name(), err)
	}

	return nil
}

func (storage *RepoStagesStorage) ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error) {
	if info, err := storage.ContainerBackend.GetImageInfo(ctx, img.Name(), container_backend.GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return false, fmt.Errorf("unable to get inspect for image %s: %w", img.Name(), err)
	} else if info != nil {
		img.SetInfo(info)
		return false, nil
	}
	return true, nil
}

func (storage *RepoStagesStorage) PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageMetadata %s %s %s %s\n", projectName, imageNameOrManagedImageName, commit, stageID)

	fullImageName := makeRepoImageMetadataName(storage.RepoAddress, imageNameOrManagedImageName, commit, stageID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageMetadata full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}
	logboek.Context(ctx).Info().LogF("Put image %s commit %s stage ID %s\n", imageNameOrManagedImageName, commit, stageID)

	return nil
}

func (storage *RepoStagesStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageMetadata %s %s %s %s\n", projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID)

	img, err := storage.selectMetadataNameImage(ctx, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID)
	if err != nil {
		return err
	}

	if img == nil {
		return nil
	}
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageMetadata full image name: %s\n", img.Tag)

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, img); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %w", img.Tag, err)
	}

	logboek.Context(ctx).Info().LogF("Removed image %s commit %s stage ID %s\n", imageNameOrManagedImageNameOrImageMetadataID, commit, stageID)

	return nil
}

func (storage *RepoStagesStorage) selectMetadataNameImage(ctx context.Context, imageNameOrManagedImageNameImageMetadataID, commit, stageID string) (*image.Info, error) {
	tryGetRepoImageFunc := func(fullImageName string) (*image.Info, error) {
		img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
		if err != nil {
			return nil, fmt.Errorf("unable to get repo image %s: %w", fullImageName, err)
		}

		return img, nil
	}

	// try as imageName or managedImageName
	{
		dockerImageName := makeRepoImageMetadataName(storage.RepoAddress, imageNameOrManagedImageNameImageMetadataID, commit, stageID)
		img, err := tryGetRepoImageFunc(dockerImageName)
		if err != nil {
			return nil, err
		}

		if img != nil {
			return img, nil
		}
	}

	// try as imageMetadataID
	{
		dockerImageTag := makeRepoImageMetadataTagNameByImageMetadataID(imageNameOrManagedImageNameImageMetadataID, commit, stageID)
		if !slug.IsValidDockerTag(dockerImageTag) { // it is not imageMetadataID
			return nil, nil
		}

		dockerImageName := makeRepoImageMetadataNameByImageMetadataID(storage.RepoAddress, imageNameOrManagedImageNameImageMetadataID, commit, stageID)
		return tryGetRepoImageFunc(dockerImageName)
	}
}

func (storage *RepoStagesStorage) IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.IsImageMetadataExist %s %s %s %s\n", projectName, imageNameOrManagedImageName, commit, stageID)

	fullImageName := makeRepoImageMetadataName(storage.RepoAddress, imageNameOrManagedImageName, commit, stageID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.IsImageMetadataExist full image name: %s\n", fullImageName)

	o := makeOptions(opts...)
	return storage.DockerRegistry.IsTagExist(ctx, fullImageName, o.dockerRegistryOptions...)
}

func (storage *RepoStagesStorage) GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImageNameStageIDCommitList %s %s\n", projectName)

	o := makeOptions(opts...)
	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	return groupImageMetadataTagsByImageName(ctx, imageNameOrManagedImageList, tags, RepoImageMetadataByCommitRecord_ImageTagPrefix)
}

func (storage *RepoStagesStorage) GetImportMetadata(ctx context.Context, _, id string) (*ImportMetadata, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadata %s\n", id)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, id)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadata full image name: %s\n", fullImageName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo image %s: %w", fullImageName, err)
	}

	if img != nil {
		return newImportMetadataFromLabels(img.Labels), nil
	}

	return nil, nil
}

func (storage *RepoStagesStorage) PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImportMetadata %v\n", metadata)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, metadata.ImportSourceID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImportMetadata full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: metadata.ToLabelsMap()}
	opts.Labels[image.WerfLabel] = projectName

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) RmImportMetadata(ctx context.Context, _, id string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImportMetadata %s\n", id)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, id)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImportMetadata full image name: %s\n", fullImageName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get repo image %s: %w", fullImageName, err)
	} else if img == nil {
		return nil
	}

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, img); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %w", img.Tag, err)
	}

	return nil
}

func (storage *RepoStagesStorage) GetImportMetadataIDs(ctx context.Context, _ string, opts ...Option) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadataIDs\n")

	o := makeOptions(opts...)
	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	var ids []string
	for _, tag := range tags {
		if !strings.HasPrefix(tag, RepoImportMetadata_ImageTagPrefix) {
			continue
		}

		ids = append(ids, getImportMetadataIDFromRepoTag(tag))
	}

	return ids, nil
}

func getImportMetadataIDFromRepoTag(tag string) string {
	return strings.TrimPrefix(tag, RepoImportMetadata_ImageTagPrefix)
}

func makeRepoImportMetadataName(repoAddress, importSourceID string) string {
	return fmt.Sprintf(RepoImportMetadata_ImageNameFormat, repoAddress, importSourceID)
}

func groupImageMetadataTagsByImageName(ctx context.Context, imageNameOrManagedImageList, tags []string, imageTagPrefix string) (map[string]map[string][]string, map[string]map[string][]string, error) {
	imageMetadataIDImageNameOrManagedImageName := map[string]string{}
	for _, imageNameOrManagedImageName := range imageNameOrManagedImageList {
		imageMetadataIDImageNameOrManagedImageName[getImageMetadataID(imageNameOrManagedImageName)] = imageNameOrManagedImageName
	}

	result := map[string]map[string][]string{}
	resultNotManagedImageName := map[string]map[string][]string{}
	for _, tag := range tags {
		var res map[string]map[string][]string

		if !strings.HasPrefix(tag, imageTagPrefix) {
			continue
		}

		sluggedImageAndCommit := strings.TrimPrefix(tag, imageTagPrefix)
		sluggedImageAndCommitParts := strings.Split(sluggedImageAndCommit, "_")
		if len(sluggedImageAndCommitParts) != 3 {
			// unexpected
			continue
		}

		tagImageNameID := sluggedImageAndCommitParts[0]
		tagCommit := sluggedImageAndCommitParts[1]
		tagStageID := sluggedImageAndCommitParts[2]

		logboek.Context(ctx).Debug().LogF("Found image ID %s commit %s stage ID %s\n", tagImageNameID, tagCommit, tagStageID)

		imageName, ok := imageMetadataIDImageNameOrManagedImageName[tagImageNameID]
		if !ok {
			res = resultNotManagedImageName
			imageName = tagImageNameID
		} else {
			res = result
		}

		stageIDCommitList, ok := res[imageName]
		if !ok {
			stageIDCommitList = map[string][]string{}
		}

		stageIDCommitList[tagStageID] = append(stageIDCommitList[tagStageID], tagCommit)
		res[imageName] = stageIDCommitList
	}

	return result, resultNotManagedImageName, nil
}

func makeRepoImageMetadataName(repoAddress, imageNameOrManagedImageName, commit, stageID string) string {
	return strings.Join([]string{repoAddress, makeRepoImageMetadataTagName(imageNameOrManagedImageName, commit, stageID)}, ":")
}

func makeRepoImageMetadataNameByImageMetadataID(repoAddress, imageMetadataID, commit, stageID string) string {
	return strings.Join([]string{repoAddress, makeRepoImageMetadataTagNameByImageMetadataID(imageMetadataID, commit, stageID)}, ":")
}

func makeRepoImageMetadataTagName(imageNameOrManagedImageName, commit, stageID string) string {
	return makeRepoImageMetadataTagNameByImageMetadataID(getImageMetadataID(imageNameOrManagedImageName), commit, stageID)
}

func makeRepoImageMetadataTagNameByImageMetadataID(imageMetadataID, commit, stageID string) string {
	return fmt.Sprintf(RepoImageMetadataByCommitRecord_TagFormat, imageMetadataID, commit, stageID)
}

func (storage *RepoStagesStorage) String() string {
	return storage.RepoAddress
}

func (storage *RepoStagesStorage) Address() string {
	return storage.RepoAddress
}

func makeRepoCustomTagMetadataRecord(repoAddress, tag string) string {
	return fmt.Sprintf(RepoCustomTagMetadata_ImageNameFormat, repoAddress, slug.LimitedSlug(tag, 48))
}

func makeRepoManagedImageRecord(repoAddress, imageNameOrManagedImageName string) string {
	return fmt.Sprintf(RepoManagedImageRecord_ImageNameFormat, repoAddress, getManagedImageID(imageNameOrManagedImageName))
}

func getManagedImageID(imageNameOrManagedImageName string) string {
	imageNameOrManagedImageName = slugImageName(imageNameOrManagedImageName)
	if !slug.IsValidDockerTag(imageNameOrManagedImageName) {
		imageNameOrManagedImageName = slug.LimitedSlug(imageNameOrManagedImageName, slug.DockerTagMaxSize-len(RepoManagedImageRecord_ImageTagPrefix))
	}

	return imageNameOrManagedImageName
}

func getManagedImageNameFromManagedImageID(managedImageID string) string {
	return unslugImageName(managedImageID)
}

func getImageMetadataID(imageNameOrManagedImageName string) string {
	return util.LegacyMurmurHash(getManagedImageNameByImageNameOrManagedImage(imageNameOrManagedImageName))
}

func getManagedImageNameByImageNameOrManagedImage(imageNameOrManagedImageName string) string {
	return getManagedImageNameFromManagedImageID(getManagedImageID(imageNameOrManagedImageName))
}

func slugImageName(imageName string) string {
	res := imageName
	res = strings.ReplaceAll(res, " ", "__space__")
	res = strings.ReplaceAll(res, "/", "__slash__")
	res = strings.ReplaceAll(res, "+", "__plus__")

	if imageName == "" {
		res = NamelessImageRecordTag
	}

	return res
}

func unslugImageName(tag string) string {
	res := tag
	res = strings.ReplaceAll(res, "__space__", " ")
	res = strings.ReplaceAll(res, "__slash__", "/")
	res = strings.ReplaceAll(res, "__plus__", "+")

	if res == NamelessImageRecordTag {
		res = ""
	}

	return res
}

func (storage *RepoStagesStorage) GetClientIDRecords(ctx context.Context, projectName string, opts ...Option) ([]*ClientIDRecord, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetClientIDRecords for project %s\n", projectName)

	o := makeOptions(opts...)
	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress, o.dockerRegistryOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %w", storage.RepoAddress, err)
	}

	var res []*ClientIDRecord
	for _, tag := range tags {
		if !strings.HasPrefix(tag, RepoClientIDRecord_ImageTagPrefix) {
			continue
		}

		tagWithoutPrefix := strings.TrimPrefix(tag, RepoClientIDRecord_ImageTagPrefix)
		dataParts := strings.SplitN(util.Reverse(tagWithoutPrefix), "-", 2)
		if len(dataParts) != 2 {
			continue
		}

		clientID, timestampMillisecStr := util.Reverse(dataParts[1]), util.Reverse(dataParts[0])

		timestampMillisec, err := strconv.ParseInt(timestampMillisecStr, 10, 64)
		if err != nil {
			continue
		}

		rec := &ClientIDRecord{ClientID: clientID, TimestampMillisec: timestampMillisec}
		res = append(res, rec)

		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetClientIDRecords got clientID record: %s\n", rec)
	}

	return res, nil
}

func (storage *RepoStagesStorage) PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostClientID %s for project %s\n", rec.ClientID, projectName)

	fullImageName := fmt.Sprintf(RepoClientIDRecord_ImageNameFormat, storage.RepoAddress, rec.ClientID, rec.TimestampMillisec)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostClientID full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted new clientID %q for project %s\n", rec.ClientID, projectName)

	return nil
}

func (storage *RepoStagesStorage) PostMultiplatformImage(ctx context.Context, projectName, tag string, allPlatformsImages []*image.Info) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostMultiplatformImage by tag %s for project %s\n", tag, projectName)

	fullImageName := fmt.Sprintf("%s:%s", storage.RepoAddress, tag)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostMultiplatformImage full image name: %s\n", fullImageName)

	opts := docker_registry.ManifestListOptions{Manifests: allPlatformsImages}
	if err := storage.DockerRegistry.PushManifestList(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted manifest list %s for project %s\n", fullImageName, projectName)

	return nil
}

func (storage *RepoStagesStorage) CopyFromStorage(ctx context.Context, src StagesStorage, projectName string, stageID image.StageID, opts CopyFromStorageOptions) (*image.StageDescription, error) {
	desc, err := storage.GetStageDescription(ctx, projectName, stageID)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage %s description: %w", stageID, err)
	}
	if desc != nil {
		return desc, nil
	}

	srcRef := src.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)
	dstRef := storage.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)
	if err := storage.DockerRegistry.CopyImage(ctx, srcRef, dstRef, docker_registry.CopyImageOptions{}); err != nil {
		return nil, fmt.Errorf("unable to copy image into registry: %w", err)
	}

	desc, err = storage.GetStageDescription(ctx, projectName, stageID)
	if err != nil {
		return nil, fmt.Errorf("unable to get stage %s description: %w", stageID, err)
	}
	return desc, nil
}

func (storage *RepoStagesStorage) FilterStagesAndProcessRelatedData(ctx context.Context, stageDescriptions []*image.StageDescription, options FilterStagesAndProcessRelatedDataOptions) ([]*image.StageDescription, error) {
	return stageDescriptions, nil
}
