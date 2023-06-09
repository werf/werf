package stage

import (
	"context"

	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
)

type Interface interface {
	Name() StageName
	LogDetailedName() string

	IsEmpty(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error)

	ExpandDependencies(ctx context.Context, c Conveyor, baseEnv map[string]string) error
	FetchDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, dockerRegistry docker_registry.GenericApiInterface) error
	GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error)
	GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error)

	PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error

	PreRun(context.Context, Conveyor) error

	SetDigest(digest string)
	GetDigest() string

	SetContentDigest(contentDigest string)
	GetContentDigest() string

	SetStageImage(*StageImage)
	GetStageImage() *StageImage

	SetGitMappings([]*GitMapping)
	GetGitMappings() []*GitMapping

	SelectSuitableStage(_ context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error)

	HasPrevStage() bool
	IsStapelStage() bool

	UsesBuildContext() bool
}
