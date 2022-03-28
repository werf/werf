package stage

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	Name() StageName
	LogDetailedName() string

	IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error)

	FetchDependencies(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime, dockerRegistry docker_registry.ApiInterface) error
	GetDependencies(ctx context.Context, c Conveyor, prevImage *StageImage, prevBuiltImage *StageImage) (string, error)
	GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error)

	PrepareImage(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime, prevBuiltImage, stageImage *StageImage) error

	PreRunHook(context.Context, Conveyor) error

	SetDigest(digest string)
	GetDigest() string

	SetContentDigest(contentDigest string)
	GetContentDigest() string

	SetStageImage(*StageImage)
	GetStageImage() *StageImage

	SetGitMappings([]*GitMapping)
	GetGitMappings() []*GitMapping

	SelectSuitableStage(_ context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error)
}
