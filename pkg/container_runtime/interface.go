package container_runtime

import (
	"context"

	"github.com/werf/werf/pkg/image"

	"github.com/docker/docker/api/types"
)

type BuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type ImageInterface interface {
	Name() string
	SetName(name string)

	Pull(ctx context.Context) error
	Push(ctx context.Context) error

	// TODO: build specifics for stapel builder and dockerfile builder
	// TODO: should be under a single separate interface
	Container() Container
	BuilderContainer() BuilderContainer
	DockerfileImageBuilder() *DockerfileImageBuilder

	Build(context.Context, BuildOptions) error
	GetBuiltId() string
	TagBuiltImage(ctx context.Context) error

	Introspect(ctx context.Context) error

	SetInspect(inspect *types.ImageInspect)
	IsExistsLocally() bool

	SetStageDescription(stage *image.StageDescription)
	GetStageDescription() *image.StageDescription
}

type Container interface {
	Name() string

	UserRunCommands() []string
	UserCommitChanges() []string

	AddServiceRunCommands(commands ...string)
	AddRunCommands(commands ...string)

	RunOptions() ContainerOptions
	CommitChangeOptions() ContainerOptions
	ServiceCommitChangeOptions() ContainerOptions
}

type BuilderContainer interface {
	AddServiceRunCommands(commands ...string)
	AddRunCommands(commands ...string)

	AddVolume(volumes ...string)
	AddVolumeFrom(volumesFrom ...string)
	AddExpose(exposes ...string)
	AddEnv(envs map[string]string)
	AddLabel(labels map[string]string)
}

type ContainerOptions interface {
	AddVolume(volumes ...string)
	AddVolumeFrom(volumesFrom ...string)
	AddExpose(exposes ...string)
	AddEnv(envs map[string]string)
	AddLabel(labels map[string]string)
	AddCmd(cmd string)
	AddWorkdir(workdir string)
	AddUser(user string)
	AddEntrypoint(entrypoint string)
	AddHealthCheck(check string)
}
