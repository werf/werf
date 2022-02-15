package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/example/stringutil"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
)

const (
	RepoStage_ImageFormat = "%s:%s-%d"

	RepoManagedImageRecord_ImageTagPrefix  = "managed-image-"
	RepoManagedImageRecord_ImageNameFormat = "%s:managed-image-%s"

	RepoRejectedStageImageRecord_ImageTagSuffix  = "-rejected"
	RepoRejectedStageImageRecord_ImageNameFormat = "%s:%s-%d-rejected"

	RepoImageMetadataByCommitRecord_ImageTagPrefix = "image-metadata-by-commit-"
	RepoImageMetadataByCommitRecord_TagFormat      = "image-metadata-by-commit-%s-%s"

	RepoClientIDRecrod_ImageTagPrefix  = "client-id-"
	RepoClientIDRecrod_ImageNameFormat = "%s:client-id-%s-%d"

	UnexpectedTagFormatErrorPrefix = "unexpected tag format"
)

func getSignatureAndUniqueIDFromRepoStageImageTag(repoStageImageTag string) (string, int64, error) {
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
	Implementation string
}

func NewRepoStagesStorage(repoAddress string, containerRuntime container_runtime.ContainerRuntime, options RepoStagesStorageOptions) (*RepoStagesStorage, error) {
	implementation := options.Implementation

	dockerRegistry, err := docker_registry.NewDockerRegistry(repoAddress, implementation, options.DockerRegistryOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating docker registry accessor for repo %q: %s", repoAddress, err)
	}

	return &RepoStagesStorage{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerRuntime: containerRuntime,
	}, nil
}

func (storage *RepoStagesStorage) ConstructStageImageName(_, signature string, uniqueID int64) string {
	return fmt.Sprintf(RepoStage_ImageFormat, storage.RepoAddress, signature, uniqueID)
}

func (storage *RepoStagesStorage) GetStagesIDs(ctx context.Context, projectName string) ([]image.StageID, error) {
	var res []image.StageID

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %s", storage.RepoAddress, err)
	} else {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesBySignature fetched tags for %q: %#v\n", storage.RepoAddress, tags)

		for _, tag := range tags {
			if strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix) || strings.HasPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix) || strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			if signature, uniqueID, err := getSignatureAndUniqueIDFromRepoStageImageTag(tag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Debug().LogLn(err.Error())
					continue
				}
				return nil, err
			} else {
				res = append(res, image.StageID{Signature: signature, UniqueID: uniqueID})

				logboek.Context(ctx).Debug().LogF("Selected stage by signature %q uniqueID %d\n", signature, uniqueID)
			}
		}

		return res, nil
	}
}

func (storage *RepoStagesStorage) DeleteStage(ctx context.Context, stageDescription *image.StageDescription, _ DeleteImageOptions) error {
	if err := storage.DockerRegistry.DeleteRepoImage(ctx, stageDescription.Info); err != nil {
		return fmt.Errorf("unable to remove repo image %s: %s", stageDescription.Info.Name, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, stageDescription.StageID.Signature, stageDescription.StageID.UniqueID)
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

	if err := storage.DockerRegistry.PushImage(ctx, rejectedImageName, docker_registry.PushImageOptions{}); err != nil {
		return fmt.Errorf("unable to push rejected stage image record %s: %s", rejectedImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Rejected stage by digest %s uniqueID %d\n", digest, uniqueID)

	return nil
}

func (storage *RepoStagesStorage) FilterStagesAndProcessRelatedData(_ context.Context, stageDescriptions []*image.StageDescription, _ FilterStagesAndProcessRelatedDataOptions) ([]*image.StageDescription, error) {
	return stageDescriptions, nil
}

func (storage *RepoStagesStorage) CreateRepo(ctx context.Context) error {
	return storage.DockerRegistry.CreateRepo(ctx, storage.RepoAddress)
}

func (storage *RepoStagesStorage) DeleteRepo(ctx context.Context) error {
	return storage.DockerRegistry.DeleteRepo(ctx, storage.RepoAddress)
}

func (storage *RepoStagesStorage) GetStagesIDsBySignature(ctx context.Context, projectName, signature string) ([]image.StageID, error) {
	var res []image.StageID

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %s", storage.RepoAddress, err)
	} else {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesBySignature fetched tags for %q: %#v\n", storage.RepoAddress, tags)

		var rejectedStages []image.StageID

		for _, tag := range tags {
			if !strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			realTag := strings.TrimSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix)

			if _, uniqueID, err := getSignatureAndUniqueIDFromRepoStageImageTag(realTag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Info().LogF("Unexpected tag %q format: %s\n", realTag, err)
					continue
				}
				return nil, fmt.Errorf("unable to get signature and uniqueID from rejected stage tag %q: %s", tag, err)
			} else {
				logboek.Context(ctx).Info().LogF("Found rejected stage %q\n", tag)
				rejectedStages = append(rejectedStages, image.StageID{Signature: signature, UniqueID: uniqueID})
			}
		}

	FindSuitableStages:
		for _, tag := range tags {
			if !strings.HasPrefix(tag, signature) {
				logboek.Context(ctx).Debug().LogF("Discard tag %q: should have prefix %q\n", tag, signature)
				continue
			}
			if strings.HasSuffix(tag, RepoRejectedStageImageRecord_ImageTagSuffix) {
				continue
			}

			if _, uniqueID, err := getSignatureAndUniqueIDFromRepoStageImageTag(tag); err != nil {
				if isUnexpectedTagFormatError(err) {
					logboek.Context(ctx).Debug().LogLn(err.Error())
					logboek.Context(ctx).Info().LogF("Unexpected tag %q format: %s\n", tag, err)
					continue
				}
				return nil, fmt.Errorf("unable to get signature and uniqueID from tag %q: %s", tag, err)
			} else {
				stageID := image.StageID{Signature: signature, UniqueID: uniqueID}

				for _, rejectedStage := range rejectedStages {
					if rejectedStage.Signature == stageID.Signature && rejectedStage.UniqueID == stageID.UniqueID {
						logboek.Context(ctx).Info().LogF("Discarding rejected stage %q\n", tag)
						continue FindSuitableStages
					}
				}

				logboek.Context(ctx).Debug().LogF("Stage %q is suitable for signature %q\n", tag, signature)
				res = append(res, stageID)
			}
		}
	}

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetRepoImagesBySignature result for %q: %#v\n", storage.RepoAddress, res)

	return res, nil
}

func (storage *RepoStagesStorage) GetStageDescription(ctx context.Context, projectName, signature string, uniqueID int64) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, signature, uniqueID)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage GetStageDescription %s %s %d\n", projectName, signature, uniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage stageImageName = %q\n", stageImageName)

	imgInfo, err := storage.DockerRegistry.GetRepoImage(ctx, stageImageName)

	if docker_registry.IsNameUnknownError(err) || docker_registry.IsManifestUnknownError(err) {
		return nil, nil
	}

	if docker_registry.IsBlobUnknownError(err) || docker_registry.IsHarbor404Error(err) {
		return nil, ErrBrokenImage
	}

	if err != nil {
		return nil, fmt.Errorf("unable to inspect repo image %s: %s", stageImageName, err)
	}

	rejectedImageName := makeRepoRejectedStageImageRecord(storage.RepoAddress, signature, uniqueID)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetStageDescription check rejected image name: %s\n", rejectedImageName)

	if rejectedImgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, rejectedImageName); err != nil {
		return nil, fmt.Errorf("unable to get repo image %q: %s", rejectedImageName, err)
	} else if rejectedImgInfo != nil {
		logboek.Context(ctx).Info().LogF("Stage signature %s uniqueID %d image is rejected: ignore stage image\n", signature, uniqueID)
		return nil, nil
	}

	return &image.StageDescription{
		StageID: &image.StageID{Signature: signature, UniqueID: uniqueID},
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

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, docker_registry.PushImageOptions{}); err != nil {
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

func (storage *RepoStagesStorage) PutImageCommit(ctx context.Context, projectName, imageName, commit string, metadata *ImageMetadata) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageCommit %s %s %s %#v\n", projectName, imageName, commit, metadata)

	fullImageName := makeRepoImageMetadataByCommitImageRecord(storage.RepoAddress, imageName, commit)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PutImageCommit full image name: %s\n", fullImageName)

	opts := docker_registry.PushImageOptions{
		Labels: map[string]string{"ContentSignature": metadata.ContentSignature},
	}
	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, opts); err != nil {
		return fmt.Errorf("unable to push image %s with metadata: %s", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Put content-signature %q into metadata for image %q by commit %s\n", metadata.ContentSignature, imageName, commit)

	return nil
}

func (storage *RepoStagesStorage) RmImageCommit(ctx context.Context, projectName, imageName, commit string) error {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageCommit %s %s %s\n", projectName, imageName, commit)

	fullImageName := makeRepoImageMetadataByCommitImageRecord(storage.RepoAddress, imageName, commit)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.RmImageCommit full image name: %s\n", fullImageName)

	if img, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName); err != nil {
		return fmt.Errorf("unable to get repo image %s: %s", fullImageName, err)
	} else if img != nil {
		if err := storage.DockerRegistry.DeleteRepoImage(ctx, img); err != nil {
			return fmt.Errorf("unable to remove repo image %s: %s", fullImageName, err)
		}

		logboek.Context(ctx).Info().LogF("Removed image %q metadata by commit %s\n", imageName, commit)
	}

	return nil
}

func (storage *RepoStagesStorage) GetImageMetadataByCommit(ctx context.Context, projectName, imageName, commit string) (*ImageMetadata, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImageStagesSignatureByCommit %s %s %s\n", projectName, imageName, commit)

	fullImageName := makeRepoImageMetadataByCommitImageRecord(storage.RepoAddress, imageName, commit)
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImageStagesSignatureByCommit full image name: %s\n", fullImageName)

	if imgInfo, err := storage.DockerRegistry.TryGetRepoImage(ctx, fullImageName); err != nil {
		return nil, fmt.Errorf("unable to get repo image %s: %s", fullImageName, err)
	} else if imgInfo != nil && imgInfo.Labels != nil {
		metadata := &ImageMetadata{ContentSignature: imgInfo.Labels["ContentSignature"]}

		logboek.Context(ctx).Debug().LogF("Got content-signature %q from image %q metadata by commit %s\n", metadata.ContentSignature, imageName, commit)

		return metadata, nil
	} else {
		logboek.Context(ctx).Debug().LogF("imgInfo = %v\n", imgInfo)
		if imgInfo != nil {
			logboek.Context(ctx).Debug().LogF("imgInfo.Labels = %v\n", imgInfo.Labels)
		}

		logboek.Context(ctx).Info().LogF("No metadata found for image %q by commit %s\n", imageName, commit)
		return nil, nil
	}
}

func (storage *RepoStagesStorage) GetImageCommits(ctx context.Context, projectName, imageName string) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.GetImageCommits %s %s\n", projectName, imageName)

	var res []string

	if tags, err := storage.DockerRegistry.Tags(ctx, storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to get repo %s tags: %s", storage.RepoAddress, err)
	} else {
		for _, tag := range tags {
			if !strings.HasPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix) {
				continue
			}

			sluggedImageAndCommit := strings.TrimPrefix(tag, RepoImageMetadataByCommitRecord_ImageTagPrefix)
			sluggedImageAndCommitParts := strings.Split(sluggedImageAndCommit, "-")
			if len(sluggedImageAndCommitParts) < 2 {
				// unexpected
				continue
			}

			commit := sluggedImageAndCommitParts[len(sluggedImageAndCommitParts)-1]
			if slugRepoImageMetadataByCommitImageRecordTag(imageName, commit) == tag {
				logboek.Context(ctx).Debug().LogF("Found image %q metadata by commit %s, full image name: %s:%s\n", imageName, commit, storage.RepoAddress, tag)
				res = append(res, commit)
			}
		}
	}

	return res, nil
}

func makeRepoImageMetadataByCommitImageRecord(repoAddress, imageName, commit string) string {
	return strings.Join([]string{
		repoAddress,
		slugRepoImageMetadataByCommitImageRecordTag(imageName, commit),
	}, ":")
}

func (storage *RepoStagesStorage) String() string {
	return fmt.Sprintf("repo stages storage (%q)", storage.RepoAddress)
}

func (storage *RepoStagesStorage) Address() string {
	return storage.RepoAddress
}

func makeRepoManagedImageRecord(repoAddress, imageName string) string {
	return fmt.Sprintf(RepoManagedImageRecord_ImageNameFormat, repoAddress, slugImageNameAsDockerImageTag(imageName))
}

func slugRepoImageMetadataByCommitImageRecordTag(imageName string, commit string) string {
	return slugImageMetadataByCommitImageRecordTag(RepoImageMetadataByCommitRecord_TagFormat, imageName, commit)
}

func slugLocalImageMetadataByCommitImageRecordTag(imageName string, commit string) string {
	return slugImageMetadataByCommitImageRecordTag(LocalImageMetadataByCommitRecord_TagFormat, imageName, commit)
}

func slugImageMetadataByCommitImageRecordTag(tagFormat string, imageName string, commit string) string {
	formattedImageName := slugImageNameAsDockerImageTag(imageName)

	tag := fmt.Sprintf(tagFormat, formattedImageName, commit)
	if len(tag) <= 128 {
		return tag
	} else {
		extraSize := len(tag) - 128
		formattedImageName = slug.LimitedSlug(formattedImageName, len(formattedImageName)-extraSize)
		formattedImageName = strings.ReplaceAll(formattedImageName, "-", "_")
		return fmt.Sprintf(tagFormat, formattedImageName, commit)
	}
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
			if !strings.HasPrefix(tag, RepoClientIDRecrod_ImageTagPrefix) {
				continue
			}

			tagWithoutPrefix := strings.TrimPrefix(tag, RepoClientIDRecrod_ImageTagPrefix)
			dataParts := strings.SplitN(stringutil.Reverse(tagWithoutPrefix), "-", 2)
			if len(dataParts) != 2 {
				continue
			}

			clientID, timestampMillisecStr := stringutil.Reverse(dataParts[1]), stringutil.Reverse(dataParts[0])

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

	fullImageName := fmt.Sprintf(RepoClientIDRecrod_ImageNameFormat, storage.RepoAddress, rec.ClientID, rec.TimestampMillisec)

	logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.PostClientID full image name: %s\n", fullImageName)

	if isExists, err := storage.DockerRegistry.IsRepoImageExists(ctx, fullImageName); err != nil {
		return err
	} else if isExists {
		logboek.Context(ctx).Debug().LogF("-- RepoStagesStorage.AddManagedImage record %q is exists => exiting\n", fullImageName)
		return nil
	}

	if err := storage.DockerRegistry.PushImage(ctx, fullImageName, docker_registry.PushImageOptions{}); err != nil {
		return fmt.Errorf("unable to push image %s: %s", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted new clientID %q for project %s\n", rec.ClientID, projectName)

	return nil
}
