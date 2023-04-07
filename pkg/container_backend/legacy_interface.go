package container_backend

import (
	"context"

	"github.com/werf/werf/pkg/image"
)

type LegacyImageInterface interface {
	Name() string
	SetName(name string)

	GetTargetPlatform() string

	Pull(ctx context.Context) error
	Push(ctx context.Context) error

	// TODO: build specifics for stapel builder and dockerfile builder
	// TODO: should be under a single separate interface
	Container() LegacyContainer
	BuilderContainer() LegacyBuilderContainer
	SetCommitChangeOptions(opts LegacyCommitChangeOptions)

	Build(context.Context, BuildOptions) error
	SetBuiltID(builtID string)
	BuiltID() string

	Introspect(ctx context.Context) error

	SetInfo(info *image.Info)

	IsExistsLocally() bool

	SetStageDescription(stage *image.StageDescription)
	GetStageDescription() *image.StageDescription

	GetFinalStageDescription() *image.StageDescription
	SetFinalStageDescription(stage *image.StageDescription)

	GetCopy() LegacyImageInterface
}

type LegacyContainer interface {
	Name() string

	UserRunCommands() []string
	UserCommitChanges() []string

	AddServiceRunCommands(commands ...string)
	AddRunCommands(commands ...string)

	RunOptions() LegacyContainerOptions
	CommitChangeOptions() LegacyContainerOptions
	ServiceCommitChangeOptions() LegacyContainerOptions
}

type LegacyBuilderContainer interface {
	AddServiceRunCommands(commands ...string)
	AddRunCommands(commands ...string)

	AddVolume(volumes ...string)
	AddVolumeFrom(volumesFrom ...string)
	AddExpose(exposes ...string)
	AddEnv(envs map[string]string)
	AddLabel(labels map[string]string)
}

type LegacyContainerOptions interface {
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
