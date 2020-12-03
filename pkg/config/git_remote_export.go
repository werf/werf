package config

import "github.com/werf/werf/pkg/giterminism_inspector"

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
	if isBranch && !giterminism_inspector.LooseGiterminism {
		// FIXME: move this message to the giterminism_inspector package

		msg := `Pay attention, werf uses git repository history to calculate stages digests. Thus, the usage of remote git mapping with branch (by default, it is master branch) might break the reproducibility of previous builds. New commits in the branch will make previously built stages not usable.

* Previous pipeline jobs (e.g. deploy) might not be retried without the image rebuild after git remote branch is modified.
* If git remote branch is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following pipeline jobs will be failed due to changing of stages digests alongside the branch HEAD.

If you want to use the branch for remote git mapping instead of commit or tag, then loose some restrictions of the werf giterminism mode with option --loose-giterminism (or WERF_LOOSE_GITERMINISM=1 env var). However it is NOT RECOMMENDED to use the remote git mapping in a such way. Use a particular unchangeable reference, tag, or commit to provide controllable and predictable lifecycle of software.`

		msg = "\n\n" + msg

		return newDetailedConfigError(msg, c.raw, c.raw.rawStapelImage.doc)
	}

	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawStapelImage.doc)
	}

	return nil
}
