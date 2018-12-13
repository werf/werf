package config

type StageDependencies struct {
	Install     []string
	Setup       []string
	BeforeSetup []string

	raw *rawStageDependencies
}

func (c *StageDependencies) validate() error {
	if !allRelativePaths(c.Install) {
		return newDetailedConfigError("`install: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	} else if !allRelativePaths(c.Setup) {
		return newDetailedConfigError("`setup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	} else if !allRelativePaths(c.BeforeSetup) {
		return newDetailedConfigError("`beforeSetup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	}
	return nil
}
