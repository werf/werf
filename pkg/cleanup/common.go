package cleanup

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type CommonOptions struct {
	DryRun bool
}

func werfDimgstagesFlushByCacheVersion(filterSet filters.Args, options CommonOptions) error {
	werfCacheVersionLabel := fmt.Sprintf("%s=%s", build.WerfCacheVersionLabel, build.BuildCacheVersion)
	filterSet.Add("label", werfCacheVersionLabel)
	images, err := werfImagesByFilterSet(filters.NewArgs())
	if err != nil {
		return err
	}

	var imagesToDelete []types.ImageSummary
	for _, img := range images {
		version, ok := img.Labels[build.WerfCacheVersionLabel]
		if !ok || version != build.BuildCacheVersion {
			imagesToDelete = append(imagesToDelete, img)
		}
	}

	if err := imagesRemove(imagesToDelete, options); err != nil {
		return err
	}

	return nil
}

func werfImagesFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	images, err := werfImagesByFilterSet(filterSet)
	if err != nil {
		return err
	}

	if err := imagesRemove(images, options); err != nil {
		return err
	}

	return nil
}

func werfImagesByFilterSet(filterSet filters.Args) ([]types.ImageSummary, error) {
	filterSet.Add("label", "werf")
	options := types.ImageListOptions{Filters: filterSet}
	return docker.Images(options)
}

func werfContainersFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	containers, err := werfContainersByFilterSet(filterSet)
	if err != nil {
		return err
	}

	if err := containersRemove(containers, options); err != nil {
		return err
	}

	return nil
}

func werfContainersByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	filterSet.Add("name", image.StageContainerNamePrefix)
	return containersByFilterSet(filterSet)
}

func containersByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Quiet = true
	containersOptions.Filters = filterSet

	return docker.Containers(containersOptions)
}

func imagesRemove(images []types.ImageSummary, options CommonOptions) error {
	var err error
	images, err = ignoreUsedImages(images)
	if err != nil {
		return err
	}

	var imageReferences []string
	for _, img := range images {
		if len(img.RepoTags) == 0 {
			imageReferences = append(imageReferences, img.ID)
		} else {
			for ind, repoTag := range img.RepoTags {
				isDanglingImage := repoTag == "<none>:<none>"
				isTaglessImage := !isDanglingImage && strings.HasSuffix(repoTag, "<none>")

				if isDanglingImage {
					imageReferences = append(imageReferences, img.ID)
				} else if isTaglessImage {
					imageReferences = append(imageReferences, img.RepoDigests[ind])
				} else {
					imageReferences = append(imageReferences, repoTag)
				}
			}
		}
	}

	if err := imageReferencesRemove(imageReferences, options); err != nil {
		return err
	}

	return nil
}

func ignoreUsedImages(images []types.ImageSummary) ([]types.ImageSummary, error) {
	filterSet := filters.NewArgs()
	for _, img := range images {
		filterSet.Add("ancestor", img.ID)
	}

	containers, err := containersByFilterSet(filterSet)
	if err != nil {
		return nil, err
	}

	var imagesToExclude []types.ImageSummary
	for _, container := range containers {
		for _, img := range images {
			if img.ID == container.ImageID {
				fmt.Printf("Skip image '%s' (used by container '%s')\n", img.ID, container.ID)
				imagesToExclude = append(imagesToExclude, img)
			}
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

func containersRemove(containers []types.Container, options CommonOptions) error {
	for _, container := range containers {
		if options.DryRun {
			fmt.Println(container.ID)
			fmt.Println()
		} else {
			if err := docker.ContainerRemove(container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				return err
			}
		}
	}

	return nil
}

func imageReferencesRemove(references []string, options CommonOptions) error {
	if len(references) != 0 {
		if options.DryRun {
			fmt.Printf(strings.Join(references, "\n"))
			fmt.Println()
		} else {
			var args []string
			args = append(args, "--force")
			args = append(args, references...)

			if err := docker.CliRmi(args...); err != nil {
				return err
			}
		}
	}

	return nil
}
