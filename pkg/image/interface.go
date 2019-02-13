package image

import "github.com/docker/docker/api/types"

type BuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type ImageInterface interface {
	Name() string
	Inspect() *types.ImageInspect
	Labels() map[string]string
	ID() string
	MustGetId() (string, error)

	Container() Container
	BuilderContainer() BuilderContainer

	IsExists() bool

	SyncDockerState() error

	Pull() error
	Untag() error

	SaveInCache() error

	Build(BuildOptions) error
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
	AddCmd(cmds ...string)
	AddOnbuild(onbuilds ...string)
	AddWorkdir(workdir string)
	AddUser(user string)
	AddEntrypoint(entrypoints ...string)
	AddStopSignal(signal string)
	AddHealthCheck(check string)
}
