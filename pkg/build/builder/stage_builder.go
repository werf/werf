package builder

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
)

type StageBuilderAccessorInterface interface {
	StapelStageBuilder() StapelStageBuilderInterface
	NativeDockerfileStageBuilder() NativeDockerfileStageBuilderInterface
	LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface
}

type StageBuilderAccessor struct {
	ContainerRuntime       container_runtime.ContainerRuntime
	Image                  container_runtime.LegacyImageInterface
	DockerfileImageBuilder *container_runtime.DockerfileImageBuilder
}

func NewStageBuilderAccessor(containerRuntime container_runtime.ContainerRuntime, image container_runtime.LegacyImageInterface) *StageBuilderAccessor {
	return &StageBuilderAccessor{
		ContainerRuntime: containerRuntime,
		Image:            image,
	}
}

func (accessor *StageBuilderAccessor) StapelStageBuilder() StapelStageBuilderInterface {
	return NewStapelStageBuilder()
}

func (accessor *StageBuilderAccessor) LegacyStapelStageBuilder() LegacyStapelStageBuilderInterface {
	return NewLegacyStapelStageBuilder(accessor.ContainerRuntime, accessor.Image)
}

func (accessor *StageBuilderAccessor) NativeDockerfileStageBuilder() NativeDockerfileStageBuilderInterface {
	if accessor.DockerfileImageBuilder == nil {
		accessor.DockerfileImageBuilder = container_runtime.NewDockerfileImageBuilder(accessor.ContainerRuntime, accessor.Image)
	}
	return accessor.DockerfileImageBuilder
}

type StapelStageBuilderInterface interface {
	AppendPrepareContainerAction(action PrepareContainerAction)
	// FIXME(stapel-to-buildah) more needed methods
}

type PrepareContainerAction interface {
	PrepareContainer(containerRoot string) error
}

// FIXME(stapel-to-buildah): full builder imlementation
type stapelStageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
}

func NewStapelStageBuilder() *stapelStageBuilder {
	return &stapelStageBuilder{}
}

func (builder *stapelStageBuilder) AppendPrepareContainerAction(action PrepareContainerAction) {
	panic("FIXME")
}

type NativeDockerfileStageBuilderInterface interface {
	Build(ctx context.Context) error
	Cleanup(ctx context.Context) error
	SetDockerfile(dockerfile []byte)
	SetDockerfileCtxRelPath(dockerfileCtxRelPath string)
	SetTarget(target string)
	AppendBuildArgs(args ...string)
	AppendAddHost(addHost ...string)
	SetNetwork(network string)
	SetSSH(ssh string)
	AppendLabels(labels ...string)
	SetContextArchivePath(contextArchivePath string)
}

type LegacyStapelStageBuilderInterface interface {
	Container() container_runtime.LegacyContainer
	BuilderContainer() container_runtime.LegacyBuilderContainer
	Build(ctx context.Context, opts container_runtime.LegacyBuildOptions) error
}

type legacyStapelStageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
	Image            container_runtime.LegacyImageInterface
}

func NewLegacyStapelStageBuilder(containerRuntime container_runtime.ContainerRuntime, image container_runtime.LegacyImageInterface) *legacyStapelStageBuilder {
	return &legacyStapelStageBuilder{
		ContainerRuntime: containerRuntime,
		Image:            image,
	}
}

func (builder *legacyStapelStageBuilder) Container() container_runtime.LegacyContainer {
	return builder.Image.Container()
}

func (builder *legacyStapelStageBuilder) BuilderContainer() container_runtime.LegacyBuilderContainer {
	return builder.Image.BuilderContainer()
}

func (builder *legacyStapelStageBuilder) Build(ctx context.Context, opts container_runtime.LegacyBuildOptions) error {
	return builder.Image.Build(ctx, opts)
}
