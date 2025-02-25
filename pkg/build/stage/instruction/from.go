package instruction

import (
	"context"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
)

type From struct {
	stage.BaseStage

	BaseImageReference  string
	BaseImageRepoDigest string

	imageCacheVersion string
}

func NewFrom(baseImageReference, baseImageRepoDigest, imageCacheVersion string, opts *stage.BaseStageOptions) *From {
	return &From{
		BaseImageReference:  baseImageReference,
		BaseImageRepoDigest: baseImageRepoDigest,
		BaseStage: *stage.NewBaseStage(
			stage.StageName("FROM"),
			opts,
		),
		imageCacheVersion: imageCacheVersion,
	}
}

func (stg *From) HasPrevStage() bool {
	return false
}

func (stg *From) IsStapelStage() bool {
	return false
}

func (stg *From) UsesBuildContext() bool {
	return false
}

func (stg *From) ExpandDependencies(ctx context.Context, c stage.Conveyor, baseEnv map[string]string) error {
	return nil
}

func (stg *From) PrepareImage(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	return nil
}

func (s *From) FetchDependencies(_ context.Context, _ stage.Conveyor, _ container_backend.ContainerBackend, _ docker_registry.GenericApiInterface) error {
	return nil
}

func (s *From) PreRun(ctx context.Context, _ stage.Conveyor) error {
	logboek.Context(ctx).LogFDetails("      ref: %s\n", s.BaseImageReference)
	if s.BaseImageRepoDigest != "" {
		logboek.Context(ctx).LogFDetails("   digest: %s\n", s.BaseImageRepoDigest)
	}
	return nil
}

func (s *From) GetDependencies(ctx context.Context, c stage.Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *stage.StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	args = append(args, "BaseImageReference", s.BaseImageReference)
	if s.BaseImageRepoDigest != "" {
		args = append(args, "BaseImageRepoDigest", s.BaseImageRepoDigest)
	}
	if s.imageCacheVersion != "" {
		args = append(args, "ImageCacheVersion", s.imageCacheVersion)
	}
	return util.Sha256Hash(args...), nil
}
