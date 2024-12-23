package host_cleaning

import (
	"github.com/docker/docker/api/types"
	docker_image "github.com/docker/docker/api/types/image"
)

type CommonOptions struct {
	RmForce                       bool
	RmiForce                      bool
	SkipUsedImages                bool
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func logImageName(image docker_image.Summary) string {
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
