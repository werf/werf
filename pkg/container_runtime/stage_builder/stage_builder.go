package stage_builder

import "github.com/werf/werf/pkg/container_runtime"

type StageBuilderInterface interface {
	StapelStageBuilder() StapelStageBuilderInterface
	DockerfileStageBuilder() DockerfileStageBuilderInterface
	LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface
}

type StageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
	FromImage        container_runtime.ImageInterface
	Image            container_runtime.LegacyImageInterface // TODO: use ImageInterface

	dockerfileStageBuilder *DockerfileStageBuilder
	stapelStageBuilder     *StapelStageBuilder
}

func (stageBuilder *StageBuilder) GetDockerfileStageBuilderImplementation() *DockerfileStageBuilder {
	return stageBuilder.dockerfileStageBuilder
}

func (stageBuilder *StageBuilder) GetStapelStageBuilderImplementation() *StapelStageBuilder {
	return stageBuilder.stapelStageBuilder
}

func NewStageBuilder(containerRuntime container_runtime.ContainerRuntime, fromImage container_runtime.ImageInterface, image container_runtime.LegacyImageInterface) *StageBuilder {
	return &StageBuilder{
		ContainerRuntime: containerRuntime,
		FromImage:        fromImage,
		Image:            image,
	}
}

func (stageBuilder *StageBuilder) StapelStageBuilder() StapelStageBuilderInterface {
	if stageBuilder.stapelStageBuilder == nil {
		stageBuilder.stapelStageBuilder = NewStapelStageBuilder(stageBuilder.ContainerRuntime, stageBuilder.FromImage, stageBuilder.Image)
	}
	return stageBuilder.stapelStageBuilder
}

func (stageBuilder *StageBuilder) LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface {
	return NewLegacyStapelStageBuilder(stageBuilder.ContainerRuntime, stageBuilder.Image)
}

func (stageBuilder *StageBuilder) DockerfileStageBuilder() DockerfileStageBuilderInterface {
	if stageBuilder.dockerfileStageBuilder == nil {
		stageBuilder.dockerfileStageBuilder = NewDockerfileStageBuilder(stageBuilder.ContainerRuntime, stageBuilder.Image)
	}
	return stageBuilder.dockerfileStageBuilder
}
