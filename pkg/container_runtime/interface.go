package container_runtime

import (
	"github.com/docker/docker/api/types"
	"github.com/werf/werf/pkg/image"
)

type BuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type ImageInterface interface {
	Name() string
	SetName(name string)

	Pull() error
	Untag() error

	// TODO: build specifics for stapel builder and dockerfile builder
	// TODO: should be under a single separate interface
	Container() Container
	BuilderContainer() BuilderContainer
	DockerfileImageBuilder() *DockerfileImageBuilder

	Build(BuildOptions) error
	GetBuiltId() string
	TagBuiltImage(name string) error
	Export(name string) error

	Introspect() error

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
