package image

type StageBuilderContainer struct{ Image *Stage }

func (c *StageBuilderContainer) AddRunCommands(commands []string) {
	c.Image.Container.AddRunCommands(commands)
}

func (c *StageBuilderContainer) AddServiceRunCommands(commands []string) {
	c.Image.Container.AddServiceRunCommands(commands)
}

func (c *StageBuilderContainer) AddVolume(volumes []string) {
	c.Image.Container.RunOptions.AddVolume(volumes)
}

func (c *StageBuilderContainer) AddVolumeFrom(volumesFrom []string) {
	c.Image.Container.RunOptions.AddVolumeFrom(volumesFrom)
}

func (c *StageBuilderContainer) AddExpose(exposes []string) {
	c.Image.Container.RunOptions.AddExpose(exposes)
}

func (c *StageBuilderContainer) AddEnv(envs map[string]interface{}) {
	c.Image.Container.RunOptions.AddEnv(envs)
}

func (c *StageBuilderContainer) AddLabel(labels map[string]interface{}) {
	c.Image.Container.RunOptions.AddLabel(labels)
}
