package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/tag_strategy"
	"github.com/werf/werf/pkg/util/parallel"
)

type ImagesRepoManager struct {
	baseManager
	ImagesRepo storage.ImagesRepo
}

func newImagesRepoManager(imagesRepo storage.ImagesRepo) *ImagesRepoManager {
	return &ImagesRepoManager{
		ImagesRepo: imagesRepo,
	}
}

func (m *ImagesRepoManager) GetRepoImage(ctx context.Context, imageName, tag string) (*image.Info, error) {
	fullImageName := m.ImagesRepo.ImageRepositoryNameWithTag(imageName, tag)
	imageInfo, err := image.CommonManifestCache.GetImageInfo(ctx, m.ImagesRepo.String(), fullImageName)
	if err != nil {
		return nil, fmt.Errorf("get image manifest from cache failed: %s", err)
	}

	if imageInfo == nil {
		logboek.Context(ctx).Info().LogF("Not found %s image info in manifest cache (CACHE MISS)\n", fullImageName)
		imageInfo, err = m.ImagesRepo.GetRepoImage(ctx, imageName, tag)
		if err != nil {
			return nil, err
		}

		if label, ok := imageInfo.Labels[image.WerfTagStrategyLabel]; ok && label == string(tag_strategy.StagesSignature) {
			if err := image.CommonManifestCache.StoreImageInfo(ctx, m.ImagesRepo.String(), imageInfo); err != nil {
				return nil, fmt.Errorf("store image manifest into cache failed: %s", err)
			}
		}
	} else {
		logboek.Context(ctx).Info().LogF("Got image %s info from manifest cache (CACHE HIT)\n", fullImageName)
	}

	return imageInfo, nil
}

func (m *ImagesRepoManager) SelectRepoImages(ctx context.Context, imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error) {
	var mutex sync.Mutex
	imageRepoTags := map[string][]string{}
	if err := parallel.DoTasks(ctx, len(imageNames), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		imageName := imageNames[taskId]

		tags, err := m.ImagesRepo.GetAllImageRepoTags(ctx, imageName)
		if err != nil {
			return fmt.Errorf("error getting all image repo tags: %s", err)
		}

		mutex.Lock()
		defer mutex.Unlock()

		imageRepoTags[imageName] = tags

		return nil
	}); err != nil {
		return nil, err
	}

	imageRepoImages := map[string][]*image.Info{}
	for _, imageName := range imageNames {
		repoImages, err := m.selectImages(ctx, imageName, imageRepoTags[imageName], f)
		if err != nil {
			return nil, err
		}

		imageRepoImages[imageName] = repoImages
	}

	return imageRepoImages, nil
}

func (m *ImagesRepoManager) selectImages(ctx context.Context, imageName string, tags []string, f func(string, *image.Info, error) (bool, error)) ([]*image.Info, error) {
	var mutex sync.Mutex
	var repoImageList []*image.Info
	if err := parallel.DoTasks(ctx, len(tags), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		tag := tags[taskId]

		if !m.ImagesRepo.IsImageRepositoryTag(imageName, tag) {
			return nil
		}

		reference := m.ImagesRepo.ImageRepositoryMetaTag(imageName, tag)
		repoImage, err := m.GetRepoImage(ctx, imageName, reference)
		if err != nil && storage.IsTagNotAssociatedWithImageError(err) {
			return nil
		}

		if f != nil {
			ok, err := f(reference, repoImage, err)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}
		} else if err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		repoImageList = append(repoImageList, repoImage)

		return nil
	}); err != nil {
		return nil, err
	}

	return repoImageList, nil
}

func (m *ImagesRepoManager) ForEachDeleteRepoImage(ctx context.Context, repoImageList []*image.Info, f func(imageInfo *image.Info, err error) error) error {
	return parallel.DoTasks(ctx, len(repoImageList), parallel.DoTasksOptions{
		MaxNumberOfWorkers: m.MaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		repoImage := repoImageList[taskId]
		err := m.ImagesRepo.DeleteRepoImage(ctx, repoImage)
		return f(repoImage, err)
	})
}
