package cleanup

import (
	"fmt"
	"strings"

	"github.com/flant/dapp/pkg/docker_registry"
)

type CommonRepoOptions struct {
	Repository string   `json:"repository"`
	DimgsNames []string `json:"dimgs_names"`
	DryRun     bool     `json:"dry_run"`
}

func repoDimgImages(options CommonRepoOptions) ([]docker_registry.RepoImage, error) {
	var dimgImages []docker_registry.RepoImage

	isNamelessDimg := len(options.DimgsNames) == 0
	if isNamelessDimg {
		namelessDimgImages, err := docker_registry.ImagesByDappDimgLabel(options.Repository, "true")
		if err != nil {
			return nil, err
		}

		dimgImages = append(dimgImages, namelessDimgImages...)
	} else {
		for _, dimgName := range options.DimgsNames {
			repository := fmt.Sprintf("%s/%s", options.Repository, dimgName)
			images, err := docker_registry.ImagesByDappDimgLabel(repository, "true")
			if err != nil {
				return nil, err
			}

			dimgImages = append(dimgImages, images...)
		}
	}

	return dimgImages, nil
}

func repoDimgstageImages(options CommonRepoOptions) ([]docker_registry.RepoImage, error) {
	return docker_registry.ImagesByDappDimgLabel(options.Repository, "false")
}

func repoImagesRemove(images []docker_registry.RepoImage, options CommonRepoOptions) error {
	isGCR, err := docker_registry.IsGCR(options.Repository)
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
	if err != nil {
		return err
	}

	if err := repoReferenceRemove(reference, options); err != nil {
		return err
	}

	return nil
}

func repoReferenceRemove(reference string, options CommonRepoOptions) error {
	fmt.Println(reference)
	if !options.DryRun {
		err := docker_registry.ImageDelete(reference)
		if err != nil {
			return err
		}
	}

	return nil
}
