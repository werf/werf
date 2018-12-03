package image

type Image interface {
	Name() string
	Labels() map[string]string

	Container() Container
	BuilderContainer() BuilderContainer

	IsExists() bool

	SyncDockerState() error

	Pull() error
}

type Container interface {
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
