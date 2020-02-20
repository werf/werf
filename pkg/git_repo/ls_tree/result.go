package ls_tree

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/flant/logboek"
)

type Result struct {
	repository        *git.Repository
	tree              *object.Tree
	treePath          string
	lsTreeEntries     []*LsTreeEntry
	submodulesResults []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
}

type LsTreeEntry struct {
	Path string
	object.TreeEntry
}

func (r *Result) HashSum() string {
	if r.empty() {
		return ""
	}

	h := sha256.New()

	sort.Slice(r.lsTreeEntries, func(i, j int) bool {
		return r.lsTreeEntries[i].Path < r.lsTreeEntries[j].Path
	})

	for _, lsTreeEntry := range r.lsTreeEntries {
		h.Write([]byte(lsTreeEntry.Hash.String()))
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].treePath < r.submodulesResults[j].treePath
	})

	for _, submoduleResult := range r.submodulesResults {
		submoduleHashSum := submoduleResult.HashSum()
		if submoduleHashSum != "" {
			h.Write([]byte(submoduleHashSum))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Result) empty() bool {
	return len(r.lsTreeEntries) == 0 && len(r.submodulesResults) == 0
}

func (r *Result) LsTree(pathFilter PathFilter) (*Result, error) {
	res := &Result{
		repository:        r.repository,
		tree:              r.tree,
		treePath:          r.treePath,
		lsTreeEntries:     []*LsTreeEntry{},
		submodulesResults: []*SubmoduleResult{},
	}

	for _, lsTreeEntry := range r.lsTreeEntries {
		var entryLsTreeEntries []*LsTreeEntry
		var entrySubmodulesResults []*SubmoduleResult

		var err error
		if lsTreeEntry.Path == "" {
			isTreeMatched, shouldWalkThrough := pathFilter.ProcessDirOrSubmodulePath(lsTreeEntry.Path)
			if isTreeMatched {
				logboek.Debug.LogLn("Root tree was added")
				entryLsTreeEntries = append(entryLsTreeEntries, lsTreeEntry)
			} else if shouldWalkThrough {
				logboek.Debug.LogLn("Root tree was opened")
				entryLsTreeEntries, entrySubmodulesResults, err = lsTreeWalk(r.repository, r.tree, r.treePath, pathFilter)
				if err != nil {
					return nil, err
				}
			}
		} else {
			entryLsTreeEntries, entrySubmodulesResults, err = lsTreeEntryMatch(r.repository, r.tree, r.treePath, lsTreeEntry, pathFilter)
		}

		res.lsTreeEntries = append(res.lsTreeEntries, entryLsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, entrySubmodulesResults...)
	}

	for _, submoduleResult := range r.submodulesResults {
		sr, err := submoduleResult.LsTree(pathFilter)
		if err != nil {
			return nil, err
		}

		res.submodulesResults = append(res.submodulesResults, &SubmoduleResult{sr})
	}

	return res, nil
}
