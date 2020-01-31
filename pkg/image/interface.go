package image

import "github.com/docker/docker/api/types"

type BuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type ImageInterface interface {
	Name() string
	SetName(name string)
	Inspect() *types.ImageInspect
	Labels() map[string]string
	ID() string

	IsExists() bool

	SyncDockerState() error

	Pull() error
	Untag() error

	// TODO: build specifics for stapel builder and dockerfile builder
	// TODO: should be under a single separate interface
	Container() Container
	BuilderContainer() BuilderContainer
	DockerfileImageBuilder() *DockerfileImageBuilder

	Build(BuildOptions) error
	GetBuiltId() (string, error)
	TagBuiltImage(name string) error

	Introspect() error
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
