package container_backend

type LegacyStageImageBuilderContainer struct{ image *LegacyStageImage }

func (c *LegacyStageImageBuilderContainer) AddRunCommands(commands ...string) {
	c.image.container.AddRunCommands(commands...)
}

func (c *LegacyStageImageBuilderContainer) AddServiceRunCommands(commands ...string) {
	c.image.container.AddServiceRunCommands(commands...)
}

func (c *LegacyStageImageBuilderContainer) AddVolume(volumes ...string) {
	c.image.container.runOptions.AddVolume(volumes...)
}

func (c *LegacyStageImageBuilderContainer) AddVolumeFrom(volumesFrom ...string) {
	c.image.container.runOptions.AddVolumeFrom(volumesFrom...)
}

func (c *LegacyStageImageBuilderContainer) AddExpose(exposes ...string) {
	c.image.container.runOptions.AddExpose(exposes...)
}

func (c *LegacyStageImageBuilderContainer) AddEnv(envs map[string]string) {
	c.image.container.runOptions.AddEnv(envs)
}

func (c *LegacyStageImageBuilderContainer) AddLabel(labels map[string]string) {
	c.image.container.runOptions.AddLabel(labels)
}
