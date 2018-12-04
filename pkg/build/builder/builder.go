package builder

type Builder interface {
	IsBeforeInstallEmpty() bool
	IsInstallEmpty() bool
	IsBeforeSetupEmpty() bool
	IsSetupEmpty() bool
	IsBuildArtifactEmpty() bool
	BeforeInstall(container Container) error
	Install(container Container) error
	BeforeSetup(container Container) error
	Setup(container Container) error
	BuildArtifact(container Container) error
	BeforeInstallChecksum() string
	InstallChecksum() string
	BeforeSetupChecksum() string
	SetupChecksum() string
	BuildArtifactChecksum() string
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
