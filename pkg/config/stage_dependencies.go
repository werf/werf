package config

type StageDependencies struct {
	Install     []string
	Setup       []string
	BeforeSetup []string

	raw *rawStageDependencies
}

func (c *StageDependencies) validate() error {
	switch {
	case !allRelativePaths(c.Install):
		return newDetailedConfigError("`install: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawStapelImage.doc)
	case !allRelativePaths(c.Setup):
		return newDetailedConfigError("`setup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawStapelImage.doc)
	case !allRelativePaths(c.BeforeSetup):
		return newDetailedConfigError("`beforeSetup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawStapelImage.doc)
	}

	return nil
}
