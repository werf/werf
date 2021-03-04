package status

type Result struct {
	Index             Scope
	Worktree          Scope
	UntrackedPathList []string
}

func (r *Result) IndexWithWorktree() Scope {
	return Scope{
		PathList:   append(r.Index.PathList, r.Worktree.PathList...),
		Submodules: append(r.Index.Submodules, r.Worktree.Submodules...),
	}
}

type Scope struct {
	PathList   []string
	Submodules []submodule
}

type submodule struct {
	Path                string
	IsAdded             bool
	IsDeleted           bool
	IsModified          bool
	HasUntrackedChanges bool
	HasTrackedChanges   bool
	IsCommitChanged     bool
}
