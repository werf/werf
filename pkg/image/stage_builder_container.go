package image

type StageBuilderContainer struct{ image *Stage }

func (c *StageBuilderContainer) AddRunCommands(commands ...string) {
	c.image.container.AddRunCommands(commands...)
}

func (c *StageBuilderContainer) AddServiceRunCommands(commands ...string) {
	c.image.container.AddServiceRunCommands(commands...)
}

func (c *StageBuilderContainer) AddVolume(volumes ...string) {
	c.image.container.runOptions.AddVolume(volumes...)
}

func (c *StageBuilderContainer) AddVolumeFrom(volumesFrom ...string) {
	c.image.container.runOptions.AddVolumeFrom(volumesFrom...)
}

func (c *StageBuilderContainer) AddExpose(exposes ...string) {
	c.image.container.runOptions.AddExpose(exposes...)
}

func (c *StageBuilderContainer) AddEnv(envs map[string]string) {
	c.image.container.runOptions.AddEnv(envs)
}

func (c *StageBuilderContainer) AddLabel(labels map[string]string) {
	c.image.container.runOptions.AddLabel(labels)
}
