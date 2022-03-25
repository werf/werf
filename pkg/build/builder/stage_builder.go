package builder

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
)

type StageBuilderAccessor interface {
	StapelStageBuilder() StapelStageBuilder
	NativeDockerfileStageBuilder() NativeDockerfileStageBuilder
	LegacyStapelStageBuilder() LegacyStapelStageBuilder
}

type stageBuilderAccessor struct {
	ContainerRuntime container_runtime.ContainerRuntime
	Image            container_runtime.LegacyImageInterface
	// FIXME(stapel-to-buildah): use this instead of LegacyImageInterface
	// nativeDockerfileStageBuilder NativeDockerfileStageBuilder
}

func NewStageBuilderAccessor(containerRuntime container_runtime.ContainerRuntime, image container_runtime.LegacyImageInterface) *stageBuilderAccessor {
	return &stageBuilderAccessor{
		ContainerRuntime: containerRuntime,
		Image:            image,
	}
}

func (accessor *stageBuilderAccessor) StapelStageBuilder() StapelStageBuilder {
	return NewStapelStageBuilder()
}

func (accessor *stageBuilderAccessor) LegacyStapelStageBuilder() LegacyStapelStageBuilder {
	return NewLegacyStapelStageBuilder(accessor.ContainerRuntime, accessor.Image)
}

func (accessor *stageBuilderAccessor) NativeDockerfileStageBuilder() NativeDockerfileStageBuilder {
	return accessor.Image.DockerfileImageBuilder()
}

type StapelStageBuilder interface {
	AppendPrepareContainerAction(action PrepareContainerAction)
	// FIXME(stapel-to-buildah) more needed methods
}

type PrepareContainerAction interface {
	PrepareContainer(containerRoot string) error
}

// FIXME(stapel-to-buildah): full imlementation of new generation builder
type stapelStageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
}

func NewStapelStageBuilder() *stapelStageBuilder {
	return &stapelStageBuilder{}
}

func (builder *stapelStageBuilder) AppendPrepareContainerAction(action PrepareContainerAction) {
	panic("FIXME")
}

type NativeDockerfileStageBuilder interface {
	Build(ctx context.Context) error
	Cleanup(ctx context.Context) error
	GetBuiltId() string
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

type LegacyStapelStageBuilder interface {
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
