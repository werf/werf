package cleaning

import (
	"strings"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker_registry"
)

type CommonRepoOptions struct {
	StagesStorage string
	ImagesRepo    string
	ImagesNames   []string
	DryRun        bool
}

func repoImages(options CommonRepoOptions) ([]docker_registry.RepoImage, error) {
	var repoImages []docker_registry.RepoImage

	if err := logboek.LogProcessInline("Getting repo images", logboek.LogProcessInlineOptions{}, func() error {
		for _, imageName := range options.ImagesNames {
			namelessImage := imageName == ""

			var repository string
			if namelessImage {
				repository = options.ImagesRepo
			} else {
				repository = strings.Join([]string{strings.TrimRight(options.ImagesRepo, "/"), imageName}, "/")
			}

			images, err := docker_registry.ImagesByWerfImageLabel(repository, "true")
			if err != nil {
				return err
			}

			repoImages = append(repoImages, images...)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logboek.LogOptionalLn()

	return repoImages, nil
}

func repoImageStagesImages(options CommonRepoOptions) ([]docker_registry.RepoImage, error) {
	return docker_registry.ImagesByWerfImageLabel(options.StagesStorage, "false")
}

func repoImagesRemove(images []docker_registry.RepoImage, options CommonRepoOptions) error {
	isGCR, err := docker_registry.IsGCR(options.ImagesRepo)
	if err != nil {
		return err
	}

	for _, image := range images {
		if isGCR {
			if err := GCRImageRemove(image, options); err != nil {
				return err
			}
		} else {
			if err := repoImageRemove(image, options); err != nil {
				return err
			}
		}
	}

	return nil
}

func GCRImageRemove(image docker_registry.RepoImage, options CommonRepoOptions) error {
	reference := strings.Join([]string{image.Repository, image.Tag}, ":")
	if err := repoReferenceRemove(reference, options); err != nil {
		return err
	}

	return nil
}

func repoImageRemove(image docker_registry.RepoImage, options CommonRepoOptions) error {
	digest, err := image.Digest()
	if err != nil {
		return err
	}

	reference := strings.Join([]string{image.Repository, digest.String()}, "@")
	if err := repoReferenceRemove(reference, options); err != nil {
		return err
	}

	logboek.LogInfoF("  tag: %s\n", image.Tag)
	logboek.LogOptionalLn()

	return nil
}

func repoReferenceRemove(reference string, options CommonRepoOptions) error {
	logboek.LogLn(reference)
	if !options.DryRun {
		err := docker_registry.ImageDelete(reference)
		if err != nil {
			return err
		}
	}

	return nil
}

func exceptRepoImages(repoImages []docker_registry.RepoImage, repoImagesToExclude ...docker_registry.RepoImage) []docker_registry.RepoImage {
	var newRepoImages []docker_registry.RepoImage

Loop:
	for _, repoImage := range repoImages {
		reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
		for _, repoImageToExclude := range repoImagesToExclude {
			referenceToExclude := strings.Join([]string{repoImageToExclude.Repository, repoImageToExclude.Tag}, ":")
			if reference == referenceToExclude {
				continue Loop
			}
		}

		newRepoImages = append(newRepoImages, repoImage)
	}

	return newRepoImages
}
