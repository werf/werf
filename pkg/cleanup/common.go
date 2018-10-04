package cleanup

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/dapp/pkg/docker"
)

type CommonOptions struct {
	DryRun bool `json:"dry_run"`
	Force  bool `json:"force"`
}

func dappImagesFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	images, err := dappImagesByFilterSet(filterSet)
	if err != nil {
		return err
	}

	if err := imagesRemove(images, options); err != nil {
		return err
	}

	return nil
}

func dappImagesByFilterSet(filterSet filters.Args) ([]types.ImageSummary, error) {
	filterSet.Add("label", "dapp")
	options := types.ImageListOptions{Filters: filterSet}
	return docker.Images(options)
}

func dappContainersFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	containers, err := dappContainersByFilterSet(filterSet)
	if err != nil {
		return err
	}

	if err := containersRemove(containers, options); err != nil {
		return err
	}

	return nil
}

func dappContainersByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	filterSet.Add("name", "dapp.build.")

	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Quiet = true
	containersOptions.Filters = filterSet

	return docker.Containers(containersOptions)
}

func imagesRemove(images []types.ImageSummary, options CommonOptions) error {
	var imageReferences []string
	for _, image := range images {
		if len(image.RepoTags) == 0 {
			imageReferences = append(imageReferences, image.ID)
		} else {
			for ind, repoTag := range image.RepoTags {
				isDanglingImage := repoTag == "<none>:<none>"
				isTaglessImage := !isDanglingImage && strings.HasSuffix(repoTag, "<none>")

				if isDanglingImage {
					imageReferences = append(imageReferences, image.ID)
				} else if isTaglessImage {
					imageReferences = append(imageReferences, image.RepoDigests[ind])
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

func containersRemove(containers []types.Container, options CommonOptions) error {
	containerRemoveOptions := types.ContainerRemoveOptions{Force: options.Force}
	for _, container := range containers {
		if options.DryRun {
			fmt.Println(container.ID) // TODO
			fmt.Println()
		} else {
			if err := docker.ContainerRemove(container.ID, containerRemoveOptions); err != nil {
				return err
			}
		}
	}

	return nil
}

func imageReferencesRemove(references []string, options CommonOptions) error {
	if len(references) != 0 {
		if options.DryRun {
			fmt.Printf(strings.Join(references, "\n")) // TODO
			fmt.Println()
		} else {
			var args []string
			if options.Force {
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
