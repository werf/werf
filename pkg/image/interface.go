package image

type BuildOptions struct {
	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

type Image interface {
	Name() string
	Labels() map[string]string
	ID() string

	Container() Container
	BuilderContainer() BuilderContainer

	IsExists() bool

	SyncDockerState() error

	Pull() error

	SaveInCache() error

	Build(BuildOptions) error
}

type Container interface {
	Name() string

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
}
