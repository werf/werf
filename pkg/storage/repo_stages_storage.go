package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

const (
	RepoStage_ImageFormat = "%s:%s-%d"

	RepoManagedImageRecord_ImageTagPrefix  = "managed-image-"
	RepoManagedImageRecord_ImageNameFormat = "%s:managed-image-%s"

	RepoRejectedStageImageRecord_ImageTagSuffix  = "-rejected"
	RepoRejectedStageImageRecord_ImageNameFormat = "%s:%s-%d-rejected"

	RepoImageMetadataByCommitRecord_ImageTagPrefix = "meta-"
	RepoImageMetadataByCommitRecord_TagFormat      = "meta-%s_%s_%s"

	RepoImportMetadata_ImageTagPrefix  = "import-metadata-"
	RepoImportMetadata_ImageNameFormat = "%s:import-metadata-%s"

	RepoClientIDRecord_ImageTagPrefix  = "client-id-"
	RepoClientIDRecord_ImageNameFormat = "%s:client-id-%s-%d"

	UnexpectedTagFormatErrorPrefix = "unexpected tag format"
)

func getDigestAndUniqueIDFromRepoStageImageTag(repoStageImageTag string) (string, int64, error) {
	parts := strings.SplitN(repoStageImageTag, "-", 2)

	if len(parts) != 2 {
		return "", 0, fmt.Errorf("%s %s", UnexpectedTagFormatErrorPrefix, repoStageImageTag)
	}

	if uniqueID, err := image.ParseUniqueIDAsTimestamp(parts[1]); err != nil {
		return "", 0, fmt.Errorf("%s %s: unable to parse unique id %s as timestamp: %s", UnexpectedTagFormatErrorPrefix, repoStageImageTag, parts[1], err)
	} else {
		return parts[0], uniqueID, nil
	}
}

func isUnexpectedTagFormatError(err error) bool {
	return strings.HasPrefix(err.Error(), UnexpectedTagFormatErrorPrefix)
}

type RepoStagesStorage struct {
	RepoAddress      string
	DockerRegistry   docker_registry.DockerRegistry
	ContainerRuntime container_runtime.ContainerRuntime
}

type RepoStagesStorageOptions struct {
	docker_registry.DockerRegistryOptions
	ContainerRegistry string
}

func NewRepoStagesStorage(repoAddress string, containerRuntime container_runtime.ContainerRuntime, options RepoStagesStorageOptions) (*RepoStagesStorage, error) {
	dockerRegistry, err := docker_registry.NewDockerRegistry(repoAddress, options.ContainerRegistry, options.DockerRegistryOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating container registry accessor for repo %q: %s", repoAddress, err)
	}

	return &RepoStagesStorage{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerRuntime: containerRuntime,
	}, nil
}

func (storage *RepoStagesStorage) ConstructStageImageName(_, digest string, uniqueID int64) string {
	return fmt.Sprintf(RepoStage_ImageFormat, storage.RepoAddress, digest, uniqueID)
}

func (storage *RepoStagesStorage) GetStagesIDs(ctx context.Context, _ string) ([]image.StageID, error) {
	var res []image.StageID

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %s", storage.RepoAddress, err)
	} else {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetStagesIDs fetched tags for %q: %#v\n", storage.RepoAddress, tags)

		for _, tag := range tags {
			if len(tag) != 70 || len(strings.Split(tag, "-")) != 2 { // 2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7-1611836746968
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
				res = append(res, image.StageID{Digest: digest, UniqueID: uniqueID})

				logboek.Context(ctx).Debug().LogF("Selected stage by digest %q uniqueID %d\n", digest, uniqueID)
			}
		}

		return res, nil
	}
}

func (storage *RepoStagesStorage) ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string) error {
	return storage.DockerRegistry.MutateAndPushImage(ctx, stageDescription.Info.Name, destinationReference, mutateExportStageConfig)
}

func mutateExportStageConfig(config v1.Config) (v1.Config, error) {
	if config.Labels == nil {
		panic("unexpected condition: stage image without labels")
	}

	for name := range config.Labels {
		if strings.HasPrefix(name, image.WerfLabelPrefix) {
			delete(config.Labels, name)
		}
	}

	return config, nil
}

func (storage *RepoStagesStorage) DeleteStage(ctx context.Context, stageDescription *image.StageDescription, _ DeleteImageOptions) error {
	if err := storage.DockerRegistry.DeleteRepoImage(ctx, stageDescription.Info); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %s", stageDescription.Info.Name, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, stageDescription.StageID.Digest, stageDescription.StageID.UniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.DeleteStage full image name: %s\n", rejectedImageName)

	if rejectedImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, rejectedImageName); err != nil {
		return fmt.Errorf("unable to get rejected image record %q: %s", rejectedImageName, err)
	} else if rejectedImgInfo != nil {
		if err := storage.DockerRegistry.DeleteRepoImage(ctx, rejectedImgInfo); err != nil {
			return fmt.Errorf("unable to remove rejected image record %q: %s", rejectedImageName, err)
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

	if isExists, err := storage.DockerRegistry.IsRepoImageExists(ctx, rejectedImageName); err != nil {
		return err
	} else if isExists {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RejectStage record %q is exists => exiting\n", rejectedImageName)
		return nil
	}

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, rejectedImageName, opts); err != nil {
		return fmt.Errorf("unable to push rejected stage image record %s: %s", rejectedImageName, err)
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

func (storage *RepoStagesStorage) GetStagesIDsByDigest(ctx context.Context, _, digest string) ([]image.StageID, error) {
	var res []image.StageID

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %s", storage.RepoAddress, err)
	} else {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesByDigest fetched tags for %q: %#v\n", storage.RepoAddress, tags)

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
				return nil, fmt.Errorf("unable to get digest and uniqueID from rejected stage tag %q: %s", tag, err)
			} else {
				logboek.Context(ctx).Info().LogF("Found rejected stage %q\n", tag)
				rejectedStages = append(rejectedStages, image.StageID{Digest: digest, UniqueID: uniqueID})
			}
		}

	FindSuitableStages:
		for _, tag := range tags {
			if !strings.HasPrefix(tag, digest) {
				logboek.Context(ctx).Debug().LogF("Discard tag %q: should have prefix %q\n", tag, digest)
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
				return nil, fmt.Errorf("unable to get digest and uniqueID from tag %q: %s", tag, err)
			} else {
				stageID := image.StageID{Digest: digest, UniqueID: uniqueID}

				for _, rejectedStage := range rejectedStages {
					if rejectedStage.Digest == stageID.Digest && rejectedStage.UniqueID == stageID.UniqueID {
						logboek.Context(ctx).Info().LogF("Discarding rejected stage %q\n", tag)
						continue FindSuitableStages
					}
				}

				logboek.Context(ctx).Debug().LogF("Stage %q is suitable for digest %q\n", tag, digest)
				res = append(res, stageID)
			}
		}
	}

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesByDigest result for %q: %#v\n", storage.RepoAddress, res)

	return res, nil
}

func (storage *RepoStagesStorage) GetStageDescription(ctx context.Context, projectName, digest string, uniqueID int64) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, digest, uniqueID)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage GetStageDescription %s %s %d\n", projectName, digest, uniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage stageImageName = %q\n", stageImageName)

	imgInfo, err := storage.DockerRegistry.GetRepoImage(ctx, stageImageName)

	if docker_registry.IsNameUnknownError(err) || docker_registry.IsManifestUnknownError(err) {
		return nil, nil
	}

	if docker_registry.IsBlobUnknownError(err) {
		return nil, ErrBrokenImage
	}

	if err != nil {
		return nil, fmt.Errorf("unable to inspect repo image %s: %s", stageImageName, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, digest, uniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetStageDescription check rejected image name: %s\n", rejectedImageName)

	if rejectedImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, rejectedImageName); err != nil {
		return nil, fmt.Errorf("unable to get repo image %q: %s", rejectedImageName, err)
	} else if rejectedImgInfo != nil {
		logboek.Context(ctx).Info().LogF("Stage digest %s uniqueID %d image is rejected: ignore stage image\n", digest, uniqueID)
		return nil, nil
	}

	return &image.StageDescription{
		StageID: &image.StageID{Digest: digest, UniqueID: uniqueID},
		Info:    imgInfo,
	}, nil
}

func (storage *RepoStagesStorage) AddManagedImage(ctx context.Context, projectName, imageName string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

	if validateImageName(imageName) != nil {
		return nil
	}

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageName)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage full image name: %s\n", fullImageName)

	if isExists, err := storage.DockerRegistry.IsRepoImageExists(ctx, fullImageName); err != nil {
		return err
	} else if isExists {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage record %q is exists => exiting\n", fullImageName)
		return nil
	}

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage record %q does not exist => creating record\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %s", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) RmManagedImage(ctx context.Context, projectName, imageName string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageName)

	if imgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName); err != nil {
		return fmt.Errorf("unable to get repo image %q info: %s", fullImageName, err)
	} else if imgInfo == nil {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmManagedImage record %q does not exist => exiting\n", fullImageName)
		return nil
	} else {
		if err := storage.DockerRegistry.DeleteRepoImage(ctx, imgInfo); err != nil {
			return fmt.Errorf("unable to delete image %q from repo: %s", fullImageName, err)
		}
	}

	return nil
}

func (storage *RepoStagesStorage) GetManagedImages(ctx context.Context, projectName string) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetManagedImages %s\n", projectName)

	var res []string

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %s", storage.RepoAddress, err)
	} else {
		for _, tag := range tags {
			if !strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix) {
				continue
			}

			managedImageName := unslugDockerImageTagAsImageName(strings.TrimPrefix(tag, RepoManagedImageRecord_ImageTagPrefix))

			if validateImageName(managedImageName) != nil {
				continue
			}

			res = append(res, managedImageName)
		}
	}

	return res, nil
}

func (storage *RepoStagesStorage) FetchImage(ctx context.Context, img container_runtime.Image) error {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		if err := containerRuntime.PullImageFromRegistry(ctx, img); err != nil {
			if strings.HasSuffix(err.Error(), "unknown blob") {
				return ErrBrokenImage
			}
			return err
		}

		return nil
	default:
		// TODO: case *container_runtime.LocalHostRuntime:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) StoreImage(ctx context.Context, img container_runtime.Image) error {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		dockerImage := img.(*container_runtime.DockerImage)

		if dockerImage.Image.GetBuiltId() != "" {
			return containerRuntime.PushBuiltImage(ctx, img)
		} else {
			return containerRuntime.PushImage(ctx, img)
		}

	default:
		// TODO: case *container_runtime.LocalHostRuntime:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) ShouldFetchImage(ctx context.Context, img container_runtime.Image) (bool, error) {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:

		dockerImage := img.(*container_runtime.DockerImage)

		if inspect, err := containerRuntime.GetImageInspect(ctx, dockerImage.Image.Name()); err != nil {
			return false, fmt.Errorf("unable to get inspect for image %s: %s", dockerImage.Image.Name(), err)
		} else if inspect != nil {
			dockerImage.Image.SetInspect(inspect)
			return false, nil
		}

		return true, nil
	default:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) PutImageMetadata(ctx context.Context, projectName, imageName, commit, stageID string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageMetadata %s %s %s %s\n", projectName, imageName, commit, stageID)

	fullImageName := makeRepoImageMetadataName(storage.RepoAddress, imageName, commit, stageID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageMetadata full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %s", fullImageName, err)
	}
	logboek.Context(ctx).Info().LogF("Put image %s commit %s stage ID %s\n", imageName, commit, stageID)

	return nil
}

func (storage *RepoStagesStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrID, commit, stageID string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageMetadata %s %s %s %s\n", projectName, imageNameOrID, commit, stageID)

	img, err := storage.selectMetadataNameImage(ctx, imageNameOrID, commit, stageID)
	if err != nil {
		return err
	}

	if img == nil {
		return nil
	}
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageMetadata full image name: %s\n", img.Tag)

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, img); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %s", img.Tag, err)
	}

	logboek.Context(ctx).Info().LogF("Removed image %s commit %s stage ID %s\n", imageNameOrID, commit, stageID)

	return nil
}

func (storage *RepoStagesStorage) selectMetadataNameImage(ctx context.Context, imageNameOrID, commit, stageID string) (*image.Info, error) {
	for _, fullImageName := range []string{
		makeRepoImageMetadataName(storage.RepoAddress, imageNameOrID, commit, stageID),
		makeRepoImageMetadataNameByImageID(storage.RepoAddress, imageNameOrID, commit, stageID),
	} {
		if img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName); err != nil {
			return nil, fmt.Errorf("unable to get repo image %s: %s", fullImageName, err)
		} else if img != nil {
			return img, nil
		}
	}

	return nil, nil
}

func (storage *RepoStagesStorage) IsImageMetadataExist(ctx context.Context, projectName, imageName, commit, stageID string) (bool, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.IsImageMetadataExist %s %s %s %s\n", projectName, imageName, commit, stageID)

	fullImageName := makeRepoImageMetadataName(storage.RepoAddress, imageName, commit, stageID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.IsImageMetadataExist full image name: %s\n", fullImageName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	return img != nil, err
}

func (storage *RepoStagesStorage) GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameList []string) (map[string]map[string][]string, map[string]map[string][]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImageNameStageIDCommitList %s %s\n", projectName)

	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get repo %s tags: %s", storage.RepoAddress, err)
	}

	return groupImageMetadataTagsByImageName(ctx, imageNameList, tags, RepoImageMetadataByCommitRecord_ImageTagPrefix)
}

func (storage *RepoStagesStorage) GetImportMetadata(ctx context.Context, _, id string) (*ImportMetadata, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadata %s\n", id)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, id)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadata full image name: %s\n", fullImageName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo image %s: %s", fullImageName, err)
	} else if img != nil {
		return newImportMetadataFromLabels(img.Labels), nil
	} else {
		return nil, nil
	}
}

func (storage *RepoStagesStorage) PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImportMetadata %v\n", metadata)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, metadata.ImportSourceID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImportMetadata full image name: %s\n", fullImageName)

	opts := &docker_registry.PushImageOptions{
		Labels: metadata.ToLabels(),
	}
	opts.Labels[image.WerfLabel] = projectName

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %s", fullImageName, err)
	}

	return nil
}

func (storage *RepoStagesStorage) RmImportMetadata(ctx context.Context, _, id string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImportMetadata %s\n", id)

	fullImageName := makeRepoImportMetadataName(storage.RepoAddress, id)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImportMetadata full image name: %s\n", fullImageName)

	img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName)
	if err != nil {
		return fmt.Errorf("unable to get repo image %s: %s", fullImageName, err)
	} else if img == nil {
		return nil
	}

	if err := storage.DockerRegistry.DeleteRepoImage(ctx, img); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %s", img.Tag, err)
	}

	return nil
}

func (storage *RepoStagesStorage) GetImportMetadataIDs(ctx context.Context, _ string) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImportMetadataIDs\n")

	tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %s", storage.RepoAddress, err)
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

func groupImageMetadataTagsByImageName(ctx context.Context, imageNameList []string, tags []string, imageTagPrefix string) (map[string]map[string][]string, map[string]map[string][]string, error) {
	imageNameNameByID := map[string]string{}
	for _, imageName := range imageNameList {
		imageNameNameByID[imageNameID(imageName)] = imageName
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

		imageName, ok := imageNameNameByID[tagImageNameID]
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

		commitList, ok := stageIDCommitList[tagStageID]
		if !ok {
			commitList = []string{}
		}

		commitList = append(commitList, tagCommit)
		stageIDCommitList[tagStageID] = commitList
		res[imageName] = stageIDCommitList
	}

	return result, resultNotManagedImageName, nil
}

func makeRepoImageMetadataName(repoAddress, imageNameOrID, commit, stageID string) string {
	return makeRepoImageMetadataNameByImageID(repoAddress, imageNameID(imageNameOrID), commit, stageID)
}

func makeRepoImageMetadataNameByImageID(repoAddress, imageID, commit, stageID string) string {
	return strings.Join([]string{
		repoAddress,
		fmt.Sprintf(RepoImageMetadataByCommitRecord_TagFormat, imageID, commit, stageID),
	}, ":")
}

func (storage *RepoStagesStorage) String() string {
	return storage.RepoAddress
}

func (storage *RepoStagesStorage) Address() string {
	return storage.RepoAddress
}

func makeRepoManagedImageRecord(repoAddress, imageName string) string {
	return fmt.Sprintf(RepoManagedImageRecord_ImageNameFormat, repoAddress, slugImageNameAsDockerImageTag(imageName))
}

func imageNameID(imageName string) string {
	return util.MurmurHash(imageName)
}

func slugImageNameAsDockerImageTag(imageName string) string {
	res := imageName
	res = strings.ReplaceAll(res, "/", "__slash__")
	res = strings.ReplaceAll(res, "+", "__plus__")

	if imageName == "" {
		res = NamelessImageRecordTag
	}

	return res
}

func unslugDockerImageTagAsImageName(tag string) string {
	res := tag
	res = strings.ReplaceAll(res, "__slash__", "/")
	res = strings.ReplaceAll(res, "__plus__", "+")

	if res == NamelessImageRecordTag {
		res = ""
	}

	return res
}

func validateImageName(name string) error {
	if strings.ToLower(name) != name {
		return fmt.Errorf("no upcase symbols allowed")
	}
	return nil
}

func (storage *RepoStagesStorage) GetClientIDRecords(ctx context.Context, projectName string) ([]*ClientIDRecord, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetClientIDRecords for project %s\n", projectName)

	var res []*ClientIDRecord

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %s", storage.RepoAddress, err)
	} else {
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
	}

	return res, nil
}

func (storage *RepoStagesStorage) PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostClientID %s for project %s\n", rec.ClientID, projectName)

	fullImageName := fmt.Sprintf(RepoClientIDRecord_ImageNameFormat, storage.RepoAddress, rec.ClientID, rec.TimestampMillisec)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostClientID full image name: %s\n", fullImageName)

	if isExists, err := storage.DockerRegistry.IsRepoImageExists(ctx, fullImageName); err != nil {
		return err
	} else if isExists {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage record %q is exists => exiting\n", fullImageName)
		return nil
	}

	opts := &docker_registry.PushImageOptions{Labels: map[string]string{image.WerfLabel: projectName}}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s: %s", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted new clientID %q for project %s\n", rec.ClientID, projectName)

	return nil
}
