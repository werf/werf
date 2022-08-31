package status

import "github.com/werf/werf/pkg/util"

type Result struct {
	Index             Index
	Worktree          Worktree
	UntrackedPathList []string
}

func (r *Result) IndexWithWorktree() Scope {
	return ComplexScope{scopes: []Scope{r.Index, r.Worktree}}
}

// PathListWithSubmodules returns a list of changed files
func (r *Result) PathListWithSubmodules() (result []string) {
	keys := map[string]bool{}

	var allPaths []string
	{
		allPaths = append(allPaths, r.IndexWithWorktree().PathList()...)
		for _, s := range r.IndexWithWorktree().Submodules() {
			allPaths = append(allPaths, s.Path)
		}

		allPaths = append(allPaths, r.UntrackedPathList...)
	}

	for _, path := range allPaths {
		_, exist := keys[path]
		if !exist {
			result = append(result, path)
			keys[path] = true
		}
	}

	return
}

type Scope interface {
	PathList() []string
	Submodules() []submodule
}

type ComplexScope struct {
	scopes []Scope
}

func (s ComplexScope) PathList() []string {
	var result []string
	for _, scope := range s.scopes {
		result = util.AddNewStringsToStringArray(result, scope.PathList()...)
	}
	return result
}

func (s ComplexScope) Submodules() []submodule {
	var result []submodule
	for _, scope := range s.scopes {
		result = append(result, scope.Submodules()...)
	}
	return result
}

type Index struct {
	baseScope
	checksum string
}

func (s Index) Checksum() string {
	return s.checksum
}

func (s *Index) addToChecksum(a ...string) {
	s.checksum = util.LegacyMurmurHash(append([]string{s.checksum}, a...)...)
}

type Worktree struct {
	baseScope
}

type baseScope struct {
	pathList   []string
	submodules []submodule
}

func (s baseScope) PathList() []string {
	return s.pathList
}

func (s baseScope) Submodules() []submodule {
	return s.submodules
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
