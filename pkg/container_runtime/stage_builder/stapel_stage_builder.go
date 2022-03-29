package stage_builder

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
)

type StapelStageBuilderInterface interface {
	AddLabels(map[string]string) StapelStageBuilderInterface
	AddBuildVolumes(volumes ...string) StapelStageBuilderInterface
	AddPrepareContainerActions(action ...container_runtime.PrepareContainerAction) StapelStageBuilderInterface
	AddUserCommands(commands ...string) StapelStageBuilderInterface

	Build(ctx context.Context, opts container_runtime.BuildOptions) error
}

type StapelStageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
	FromImage        container_runtime.ImageInterface
	Image            container_runtime.ImageInterface

	labels                  []string
	buildVolumes            []string
	prepareContainerActions []container_runtime.PrepareContainerAction
	userCommands            []string
}

func NewStapelStageBuilder(containerRuntime container_runtime.ContainerRuntime, fromImage, image container_runtime.ImageInterface) *StapelStageBuilder {
	return &StapelStageBuilder{
		ContainerRuntime: containerRuntime,
		FromImage:        fromImage,
		Image:            image,
	}
}

func (builder *StapelStageBuilder) Build(ctx context.Context, opts container_runtime.BuildOptions) error {
	// TODO: support introspect options

	builtID, err := builder.ContainerRuntime.BuildStapelStage(ctx, builder.FromImage.Name(), container_runtime.BuildStapelStageOpts{
		BuildVolumes:            builder.buildVolumes,
		Labels:                  builder.labels,
		UserCommands:            builder.userCommands,
		PrepareContainerActions: builder.prepareContainerActions,
	})
	if err != nil {
		return fmt.Errorf("error building stapel stage with %s: %s", builder.ContainerRuntime.String(), err)
	}

	builder.Image.SetBuiltID(builtID)

	return nil
}

func (builder *StapelStageBuilder) AddLabels(labels map[string]string) StapelStageBuilderInterface {
	for k, v := range labels {
		builder.labels = append(builder.labels, fmt.Sprintf("%s=%s", k, v))
	}

	return builder
}

func (builder *StapelStageBuilder) AddBuildVolumes(volumes ...string) StapelStageBuilderInterface {
	builder.buildVolumes = append(builder.buildVolumes, volumes...)

	return builder
}

func (builder *StapelStageBuilder) AddPrepareContainerActions(actions ...container_runtime.PrepareContainerAction) StapelStageBuilderInterface {
	builder.prepareContainerActions = append(builder.prepareContainerActions, actions...)

	return builder
}

func (builder *StapelStageBuilder) AddUserCommands(commands ...string) StapelStageBuilderInterface {
	builder.userCommands = append(builder.userCommands, commands...)

	return builder
}
