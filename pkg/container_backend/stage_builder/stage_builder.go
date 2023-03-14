package stage_builder

import (
	"context"

	"github.com/werf/werf/pkg/container_backend"
)

type StageBuilderInterface interface {
	StapelStageBuilder() StapelStageBuilderInterface
	DockerfileBuilder() DockerfileBuilderInterface
	DockerfileStageBuilder() DockerfileStageBuilderInterface
	LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface

	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

func NewStageBuilder(containerBackend container_backend.ContainerBackend, baseImage string, image container_backend.LegacyImageInterface) *StageBuilder {
	return &StageBuilder{
		ContainerBackend: containerBackend,
		BaseImage:        baseImage,
		Image:            image,
	}
}

type StageBuilder struct {
	ContainerBackend container_backend.ContainerBackend
	BaseImage        string
	Image            container_backend.LegacyImageInterface // TODO: use ImageInterface

	dockerfileBuilder        *DockerfileBuilder
	dockerfileStageBuilder   *DockerfileStageBuilder
	stapelStageBuilder       *StapelStageBuilder
	legacyStapelStageBuilder *LegacyStapelStageBuilder
}

func (stageBuilder *StageBuilder) GetDockerfileBuilderImplementation() *DockerfileBuilder {
	return stageBuilder.dockerfileBuilder
}

func (stageBuilder *StageBuilder) GetDockerfileStageBuilderImplementation() *DockerfileStageBuilder {
	return stageBuilder.dockerfileStageBuilder
}

func (stageBuilder *StageBuilder) GetStapelStageBuilderImplementation() *StapelStageBuilder {
	return stageBuilder.stapelStageBuilder
}

func (stageBuilder *StageBuilder) GetLegacyStapelStageBuilderImplmentation() *LegacyStapelStageBuilder {
	return stageBuilder.legacyStapelStageBuilder
}

func (stageBuilder *StageBuilder) StapelStageBuilder() StapelStageBuilderInterface {
	if stageBuilder.stapelStageBuilder == nil {
		stageBuilder.stapelStageBuilder = NewStapelStageBuilder(stageBuilder.ContainerBackend, stageBuilder.BaseImage, stageBuilder.Image)
	}
	return stageBuilder.stapelStageBuilder
}

func (stageBuilder *StageBuilder) LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface {
	if stageBuilder.legacyStapelStageBuilder == nil {
		stageBuilder.legacyStapelStageBuilder = NewLegacyStapelStageBuilder(stageBuilder.ContainerBackend, stageBuilder.Image)
	}
	return stageBuilder.legacyStapelStageBuilder
}

func (stageBuilder *StageBuilder) DockerfileBuilder() DockerfileBuilderInterface {
	if stageBuilder.dockerfileBuilder == nil {
		stageBuilder.dockerfileBuilder = NewDockerfileBuilder(stageBuilder.ContainerBackend, stageBuilder.Image)
	}
	return stageBuilder.dockerfileBuilder
}

func (stageBuilder *StageBuilder) DockerfileStageBuilder() DockerfileStageBuilderInterface {
	if stageBuilder.dockerfileStageBuilder == nil {
		stageBuilder.dockerfileStageBuilder = NewDockerfileStageBuilder(stageBuilder.ContainerBackend, stageBuilder.BaseImage, stageBuilder.Image)
	}
	return stageBuilder.dockerfileStageBuilder
}

func (stageBuilder *StageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	switch {
	case stageBuilder.dockerfileBuilder != nil:
		return stageBuilder.dockerfileBuilder.Build(ctx, opts)
	case stageBuilder.dockerfileStageBuilder != nil:
		return stageBuilder.dockerfileStageBuilder.Build(ctx, opts)
	case stageBuilder.stapelStageBuilder != nil:
		return stageBuilder.stapelStageBuilder.Build(ctx, opts)
	case stageBuilder.legacyStapelStageBuilder != nil:
		return stageBuilder.legacyStapelStageBuilder.Build(ctx, opts)
	}

	panic("no builder has been activated yet")
}
