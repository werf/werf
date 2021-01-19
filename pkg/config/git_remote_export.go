package config

import (
	"github.com/werf/werf/pkg/giterminism_manager"
)

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Tag    string
	Commit string

	raw *rawGit
}

func (c *GitRemoteExport) validate(giterminismManager giterminism_manager.Interface) error {
	isDefaultMasterBranch := c.Branch == "" && c.Commit == "" && c.Tag == ""
	isBranch := isDefaultMasterBranch || c.Branch != ""

	if isBranch {
		if err := giterminismManager.Inspector().InspectConfigStapelGitBranch(); err != nil {
			return newDetailedConfigError(err.Error(), c.raw, c.raw.rawStapelImage.doc)
		}
	}

	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawStapelImage.doc)
	}

	return nil
}
