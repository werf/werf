package config

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Tag    string
	Commit string

	raw *rawGit
}

func (c *GitRemoteExport) validate() error {
	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawDimg.doc)
	}
	return nil
}
