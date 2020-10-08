package stage

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	Name() StageName
	LogDetailedName() string

	IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error)

	FetchDependencies(ctx context.Context, c Conveyor, cr container_runtime.ContainerRuntime) error
	GetDependencies(ctx context.Context, c Conveyor, prevImage container_runtime.ImageInterface, prevBuiltImage container_runtime.ImageInterface) (string, error)
	GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error)

	PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error

	PreRunHook(context.Context, Conveyor) error

	SetDigest(digest string)
	GetDigest() string

	SetContentDigest(contentDigest string)
	GetContentDigest() string

	SetImage(container_runtime.ImageInterface)
	GetImage() container_runtime.ImageInterface

	SetGitMappings([]*GitMapping)
	GetGitMappings() []*GitMapping

	SelectSuitableStage(_ context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error)
}
