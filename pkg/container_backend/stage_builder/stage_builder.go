package stage_builder

import "github.com/werf/werf/pkg/container_backend"

type StageBuilderInterface interface {
	StapelStageBuilder() StapelStageBuilderInterface
	DockerfileStageBuilder() DockerfileStageBuilderInterface
	LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface
}

type StageBuilder struct {
	ContainerBackend container_backend.ContainerBackend
	FromImage        container_backend.ImageInterface
	Image            container_backend.LegacyImageInterface // TODO: use ImageInterface

	dockerfileStageBuilder *DockerfileStageBuilder
	stapelStageBuilder     *StapelStageBuilder
}

func (stageBuilder *StageBuilder) GetDockerfileStageBuilderImplementation() *DockerfileStageBuilder {
	return stageBuilder.dockerfileStageBuilder
}

func (stageBuilder *StageBuilder) GetStapelStageBuilderImplementation() *StapelStageBuilder {
	return stageBuilder.stapelStageBuilder
}

func NewStageBuilder(containerBackend container_backend.ContainerBackend, fromImage container_backend.ImageInterface, image container_backend.LegacyImageInterface) *StageBuilder {
	return &StageBuilder{
		ContainerBackend: containerBackend,
		FromImage:        fromImage,
		Image:            image,
	}
}

func (stageBuilder *StageBuilder) StapelStageBuilder() StapelStageBuilderInterface {
	if stageBuilder.stapelStageBuilder == nil {
		stageBuilder.stapelStageBuilder = NewStapelStageBuilder(stageBuilder.ContainerBackend, stageBuilder.FromImage, stageBuilder.Image)
	}
	return stageBuilder.stapelStageBuilder
}

func (stageBuilder *StageBuilder) LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface {
	return NewLegacyStapelStageBuilder(stageBuilder.ContainerBackend, stageBuilder.Image)
}

func (stageBuilder *StageBuilder) DockerfileStageBuilder() DockerfileStageBuilderInterface {
	if stageBuilder.dockerfileStageBuilder == nil {
		stageBuilder.dockerfileStageBuilder = NewDockerfileStageBuilder(stageBuilder.ContainerBackend, stageBuilder.Image)
	}
	return stageBuilder.dockerfileStageBuilder
}
