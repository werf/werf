package host_cleaning

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
)

func werfImagesFlushByFilterSet(ctx context.Context, filterSet filters.Args, options CommonOptions) error {
	images, err := imagesByFilterSet(ctx, filterSet)
	if err != nil {
		return err
	}

	images, err = processUsedImages(ctx, images, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(ctx, images, options); err != nil {
		return err
	}

	return nil
}

func trueDanglingImages(ctx context.Context) ([]types.ImageSummary, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("dangling", "true")
	danglingImageList, err := imagesByFilterSet(ctx, filterSet)
	if err != nil {
		return nil, err
	}

	var trueDanglingImageList []types.ImageSummary
	for _, image := range danglingImageList {
		if len(image.RepoTags) == 0 && len(image.RepoDigests) == 0 {
			trueDanglingImageList = append(trueDanglingImageList, image)
		}
	}

	return trueDanglingImageList, nil
}

func imagesByFilterSet(ctx context.Context, filterSet filters.Args) ([]types.ImageSummary, error) {
	options := types.ImageListOptions{Filters: filterSet}
	return docker.Images(ctx, options)
}

func processUsedImages(ctx context.Context, images []types.ImageSummary, options CommonOptions) ([]types.ImageSummary, error) {
	filterSet := filters.NewArgs()
	for _, img := range images {
		filterSet.Add("ancestor", img.ID)
	}

	containers, err := containersByFilterSet(ctx, filterSet)
	if err != nil {
		return nil, err
	}

	var imagesToExclude []types.ImageSummary
	var containersToRemove []types.Container
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

	if err := containersRemove(ctx, containersToRemove, options); err != nil {
		return nil, err
	}

	for _, img := range images {
		// IMPORTANT: Prevent freshly built images, but not saved into the stages storage yet from being deleted
		if time.Since(time.Unix(img.Created, 0)) < 3*time.Hour {
			imagesToExclude = append(imagesToExclude, img)
		}
	}

	for _, img := range imagesToExclude {
		images = exceptImage(images, img)
	}

	return images, nil
}

func exceptImage(images []types.ImageSummary, imageToExclude types.ImageSummary) []types.ImageSummary {
	var newImages []types.ImageSummary
	for _, img := range images {
		if !reflect.DeepEqual(imageToExclude, img) {
			newImages = append(newImages, img)
		}
	}

	return newImages
}

func imagesRemove(ctx context.Context, images []types.ImageSummary, options CommonOptions) error {
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

	if err := imageReferencesRemove(ctx, imageReferences, options); err != nil {
		return err
	}

	return nil
}

func imageReferencesRemove(ctx context.Context, references []string, options CommonOptions) error {
	if len(references) != 0 {
		if options.DryRun {
			logboek.Context(ctx).LogLn(strings.Join(references, "\n"))
			logboek.Context(ctx).LogOptionalLn()
		} else {
			var args []string

			if options.RmiForce {
				args = append(args, "--force")
			}
			args = append(args, references...)

			if err := docker.CliRmi_LiveOutput(ctx, args...); err != nil {
				return err
			}
		}
	}

	return nil
}
