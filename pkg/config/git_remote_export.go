package config

type GitRemoteExport struct {
	*GitLocalExport
	Branch                                          string
	Tag                                             string
	Commit                                          string
	HerebyIAdmitThatBranchMightBreakReproducibility bool

	raw *rawGit
}

func (c *GitRemoteExport) validate() error {
	isDefaultMasterBranch := c.Branch == "" && c.Commit == "" && c.Tag == ""
	isBranch := isDefaultMasterBranch || c.Branch != ""
	if isBranch && !c.HerebyIAdmitThatBranchMightBreakReproducibility {
		msg := `Pay attention, werf uses git repository history to calculate stages signatures. Thus, the usage of remote git mapping with branch (by default, it is master branch) might break the reproducibility of previous builds. New commits in the branch will make previously built stages not usable.

* Previous pipeline jobs (e.g. deploy) might not be retried without the image rebuild after git remote branch is modified.
* If git remote branch is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following pipeline jobs will be failed due to changing of stages signatures alongside the branch HEAD.

If you want to use the branch for remote git mapping instead of commit or tag, add 'herebyIAdmitThatBranchMightBreakReproducibility: true' into the remote git mapping section.

We do not recommend using the remote git mapping such way. Use a particular unchangeable reference, tag, or commit to provide controllable and predictable lifecycle of software.`

		msg = "\n\n" + msg

		return newDetailedConfigError(msg, c.raw, c.raw.rawStapelImage.doc)
	}

	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawStapelImage.doc)
	}

	return nil
}
