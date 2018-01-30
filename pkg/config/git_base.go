package config

type GitBase struct {
	*GitExport
	As                string
	StageDependencies *StageDependencies

	Raw *RawGit
}

func (c *GitBase) Validate() error {
	return nil
}
