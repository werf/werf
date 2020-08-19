package builder

import (
	"context"
	"os"
)

type Builder interface {
	IsBeforeInstallEmpty(ctx context.Context) bool
	IsInstallEmpty(ctx context.Context) bool
	IsBeforeSetupEmpty(ctx context.Context) bool
	IsSetupEmpty(ctx context.Context) bool
	BeforeInstall(ctx context.Context, container Container) error
	Install(ctx context.Context, container Container) error
	BeforeSetup(ctx context.Context, container Container) error
	Setup(ctx context.Context, container Container) error
	BeforeInstallChecksum(ctx context.Context) string
	InstallChecksum(ctx context.Context) string
	BeforeSetupChecksum(ctx context.Context) string
	SetupChecksum(ctx context.Context) string
}

type Container interface {
	AddRunCommands(commands ...string)
	AddServiceRunCommands(commands ...string)
	AddVolumeFrom(volumesFrom ...string)
	AddVolume(volumes ...string)
	AddExpose(exposes ...string)
	AddEnv(envs map[string]string)
	AddLabel(labels map[string]string)
}

func debugUserStageChecksum() bool {
	return os.Getenv("WERF_DEBUG_USER_STAGE_CHECKSUM") == "1"
}
