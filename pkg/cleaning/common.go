package cleaning

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type CommonOptions struct {
	DryRun                        bool
	RmForce                       bool
	RmiForce                      bool
	SkipUsedImages                bool
	RmContainersThatUseWerfImages bool
}

func werfImagesFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	images, err := werfImagesByFilterSet(filterSet)
	if err != nil {
		return err
	}

	images, err = processUsedImages(images, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(images, options); err != nil {
		return err
	}

	return nil
}

func werfImagesByFilterSet(filterSet filters.Args) ([]types.ImageSummary, error) {
	filterSet.Add("label", image.WerfLabel)
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

func processUsedImages(images []types.ImageSummary, options CommonOptions) ([]types.ImageSummary, error) {
	filterSet := filters.NewArgs()
	for _, img := range images {
		filterSet.Add("ancestor", img.ID)
	}

	containers, err := containersByFilterSet(filterSet)
	if err != nil {
		return nil, err
	}

	var imagesToExclude []types.ImageSummary
	var containersToRemove []types.Container
	for _, container := range containers {
		for _, img := range images {
			if img.ID == container.ImageID {
				if options.SkipUsedImages {
					logboek.Default.LogFDetails("Skip image %s (used by container %s)\n", logImageName(img), logContainerName(container))
					imagesToExclude = append(imagesToExclude, img)
				} else if options.RmContainersThatUseWerfImages {
					containersToRemove = append(containersToRemove, container)
				} else {
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", logImageName(img), logContainerName(container), "Use --force option to remove all containers that are based on deleting werf docker images")
				}
			}
		}
	}

	if err := containersRemove(containersToRemove, options); err != nil {
		return nil, err
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
			logboek.LogLn(logContainerName(container))
			logboek.LogOptionalLn()
		} else {
			if err := docker.ContainerRemove(container.ID, types.ContainerRemoveOptions{Force: options.RmForce}); err != nil {
				return err
			}
		}
	}

	return nil
}

func imageReferencesRemove(references []string, options CommonOptions) error {
	if len(references) != 0 {
		if options.DryRun {
			logboek.LogLn(strings.Join(references, "\n"))
			logboek.LogOptionalLn()
		} else {
			var args []string

			if options.RmiForce {
				args = append(args, "--force")
			}
			args = append(args, references...)

			if err := docker.CliRmi(args...); err != nil {
				return err
			}
		}
	}

	return nil
}

func danglingFilterSet() filters.Args {
	filterSet := filters.NewArgs()
	filterSet.Add("dangling", "true")
	return filterSet
}

func logImageName(image types.ImageSummary) string {
	name := image.ID
	if len(image.RepoTags) != 0 {
		name = image.RepoTags[0]
	}

	return name
}

func logContainerName(container types.Container) string {
	name := container.ID
	if len(container.Names) != 0 {
		name = container.Names[0]
	}

	return name
}
