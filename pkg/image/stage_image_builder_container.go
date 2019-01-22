package image

type StageImageBuilderContainer struct{ image *StageImage }

func (c *StageImageBuilderContainer) AddRunCommands(commands ...string) {
	c.image.container.AddRunCommands(commands...)
}

func (c *StageImageBuilderContainer) AddServiceRunCommands(commands ...string) {
	c.image.container.AddServiceRunCommands(commands...)
}

func (c *StageImageBuilderContainer) AddVolume(volumes ...string) {
	c.image.container.runOptions.AddVolume(volumes...)
}

func (c *StageImageBuilderContainer) AddVolumeFrom(volumesFrom ...string) {
	c.image.container.runOptions.AddVolumeFrom(volumesFrom...)
}

func (c *StageImageBuilderContainer) AddExpose(exposes ...string) {
	c.image.container.runOptions.AddExpose(exposes...)
}

func (c *StageImageBuilderContainer) AddEnv(envs map[string]string) {
	c.image.container.runOptions.AddEnv(envs)
}

func (c *StageImageBuilderContainer) AddLabel(labels map[string]string) {
	c.image.container.runOptions.AddLabel(labels)
}
