package host_cleaning

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

func werfImagesFlushByFilterSet(ctx context.Context, backend container_backend.ContainerBackend, imagesOptions container_backend.ImagesOptions, options CommonOptions) error {
	images, err := backend.Images(ctx, imagesOptions)
	if err != nil {
		return err
	}

	images, err = processUsedImages(ctx, backend, images, options)
	if err != nil {
		return err
	}

	if err = imagesRemove(ctx, backend, images, options); err != nil {
		return err
	}

	return nil
}

func trueDanglingImages(ctx context.Context, backend container_backend.ContainerBackend) (image.ImagesList, error) {
	imagesOptions := buildImagesOptions(
		util.NewPair("dangling", "true"),
	)

	danglingImageList, err := backend.Images(ctx, imagesOptions)
	if err != nil {
		return nil, err
	}

	var trueDanglingImageList image.ImagesList
	for _, image := range danglingImageList {
		if len(image.RepoTags) == 0 && len(image.RepoDigests) == 0 {
			trueDanglingImageList = append(trueDanglingImageList, image)
		}
	}

	return trueDanglingImageList, nil
}

func processUsedImages(ctx context.Context, backend container_backend.ContainerBackend, images image.ImagesList, options CommonOptions) (image.ImagesList, error) {
	containersOptionsFilters := make([]image.ContainerFilter, len(images))
	for i, img := range images {
		containersOptionsFilters[i] = image.ContainerFilter{Ancestor: img.ID}
	}

	containers, err := backend.Containers(ctx, buildContainersOptions(containersOptionsFilters...))
	if err != nil {
		return nil, err
	}

	var imagesToExclude image.ImagesList
	var containersToRemove []image.Container
	for _, container := range containers {
		for _, img := range images {
			if img.ID == container.ImageID {
				switch {
				case options.SkipUsedImages:
					logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", logImageName(img), logContainerName(container))
					imagesToExclude = append(imagesToExclude, img)
				case options.RmContainersThatUseWerfImages:
					containersToRemove = append(containersToRemove, container)
				default:
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", logImageName(img), logContainerName(container), "Use --force option to remove all containers that are based on deleting werf docker images")
				}
			}
		}
	}

	if err = containersRemove(ctx, backend, containersToRemove, options); err != nil {
		return nil, err
	}

	for _, img := range images {
		// IMPORTANT: Prevent freshly built images, but not saved into the stages storage yet from being deleted
		if time.Since(img.Created) < 3*time.Hour {
			imagesToExclude = append(imagesToExclude, img)
		}
	}

	for _, img := range imagesToExclude {
		images = exceptImage(images, img)
	}

	return images, nil
}

func exceptImage(images image.ImagesList, imageToExclude image.Summary) image.ImagesList {
	var newImages image.ImagesList
	for _, img := range images {
		if !reflect.DeepEqual(imageToExclude, img) {
			newImages = append(newImages, img)
		}
	}

	return newImages
}

func imagesRemove(ctx context.Context, backend container_backend.ContainerBackend, images image.ImagesList, options CommonOptions) error {
	var imageReferences []string

	for _, img := range images {
		if len(img.RepoTags) == 0 {
			imageReferences = append(imageReferences, img.ID)
		} else {
			for _, repoTag := range img.RepoTags {
				isDanglingImage := repoTag == "<none>:<none>"
				isTaglessImage := !isDanglingImage && strings.HasSuffix(repoTag, "<none>")

				if isDanglingImage || isTaglessImage {
					imageReferences = append(imageReferences, img.ID)
				} else {
					imageReferences = append(imageReferences, repoTag)
				}
			}
		}
	}

	if err := imageReferencesRemove(ctx, backend, imageReferences, options); err != nil {
		return err
	}

	return nil
}

func imageReferencesRemove(ctx context.Context, backend container_backend.ContainerBackend, references []string, options CommonOptions) error {
	for _, ref := range references {
		if options.DryRun {
			logboek.Context(ctx).LogLn(ref)
			logboek.Context(ctx).LogOptionalLn()
		} else {
			err := backend.Rmi(ctx, ref, container_backend.RmiOpts{
				Force: options.RmiForce,
			})
			if err != nil {
				return fmt.Errorf("container_backend rmi: %w", err)
			}
		}
	}

	return nil
}

func buildImagesOptions(filters ...util.Pair[string, string]) container_backend.ImagesOptions {
	opts := container_backend.ImagesOptions{}
	opts.Filters = filters
	return opts
}
