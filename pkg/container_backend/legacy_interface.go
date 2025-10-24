package container_backend

import (
	"context"

	"github.com/werf/werf/v2/pkg/image"
)

//go:generate mockgen -source legacy_interface.go -package mock -destination ../../test/mock/legacy_interface.go

type LegacyImageInterface interface {
	Name() string
	SetName(name string)

	GetTargetPlatform() string

	Pull(ctx context.Context) error
	Push(ctx context.Context) error

	SetBuildServiceLabels(labels map[string]string)
	GetBuildServiceLabels() map[string]string

	// TODO: build specifics for stapel builder and dockerfile builder
	// TODO: should be under a single separate interface
	Container() LegacyContainer
	BuilderContainer() LegacyBuilderContainer
	SetCommitChangeOptions(opts LegacyCommitChangeOptions)

	Build(context.Context, BuildOptions) error
	SetBuiltID(builtID string)
	BuiltID() string
	Mutate(ctx context.Context, mutationFunc func(builtID string) (string, error)) error

	Introspect(ctx context.Context) error

	SetInfo(info *image.Info)

	IsExistsLocally() bool

	SetStageDesc(*image.StageDesc)
	GetStageDesc() *image.StageDesc

	GetFinalStageDesc() *image.StageDesc
	SetFinalStageDesc(*image.StageDesc)

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
	MountSSHAgentSocket(sshAuthSock string)
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
