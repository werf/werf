package config

import (
	"context"

	"github.com/werf/werf/pkg/giterminism_inspector"
)

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Tag    string
	Commit string

	raw *rawGit
}

func (c *GitRemoteExport) validate() error {
	isDefaultMasterBranch := c.Branch == "" && c.Commit == "" && c.Tag == ""
	isBranch := isDefaultMasterBranch || c.Branch != ""
	if isBranch {
		if err := giterminism_inspector.ReportConfigStapelGitBranch(context.Background()); err != nil {
			errMsg := "\n\n" + err.Error()
			return newDetailedConfigError(errMsg, c.raw, c.raw.rawStapelImage.doc)
		}
	}

	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawStapelImage.doc)
	}

	return nil
}
