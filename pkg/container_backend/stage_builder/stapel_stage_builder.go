package stage_builder

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_backend"
)

type StapelStageBuilderInterface interface {
	AddLabels(map[string]string) StapelStageBuilderInterface
	AddBuildVolumes(volumes ...string) StapelStageBuilderInterface
	AddPrepareContainerActions(action ...container_backend.PrepareContainerAction) StapelStageBuilderInterface
	AddUserCommands(commands ...string) StapelStageBuilderInterface

	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

type StapelStageBuilder struct {
	ContainerBackend container_backend.ContainerBackend
	FromImage        container_backend.ImageInterface
	Image            container_backend.ImageInterface

	labels                  []string
	buildVolumes            []string
	prepareContainerActions []container_backend.PrepareContainerAction
	userCommands            []string
}

func NewStapelStageBuilder(containerBackend container_backend.ContainerBackend, fromImage, image container_backend.ImageInterface) *StapelStageBuilder {
	return &StapelStageBuilder{
		ContainerBackend: containerBackend,
		FromImage:        fromImage,
		Image:            image,
	}
}

func (builder *StapelStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	// TODO: support introspect options

	builtID, err := builder.ContainerBackend.BuildStapelStage(ctx, builder.FromImage.Name(), container_backend.BuildStapelStageOpts{
		BuildVolumes:            builder.buildVolumes,
		Labels:                  builder.labels,
		UserCommands:            builder.userCommands,
		PrepareContainerActions: builder.prepareContainerActions,
	})
	if err != nil {
		return fmt.Errorf("error building stapel stage with %s: %w", builder.ContainerBackend.String(), err)
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

func (builder *StapelStageBuilder) AddPrepareContainerActions(actions ...container_backend.PrepareContainerAction) StapelStageBuilderInterface {
	builder.prepareContainerActions = append(builder.prepareContainerActions, actions...)

	return builder
}

func (builder *StapelStageBuilder) AddUserCommands(commands ...string) StapelStageBuilderInterface {
	builder.userCommands = append(builder.userCommands, commands...)

	return builder
}
