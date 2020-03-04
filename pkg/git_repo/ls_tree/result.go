package ls_tree

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/path_matcher"
)

type Result struct {
	repository                       *git.Repository
	tree                             *object.Tree
	treeFilepath                     string
	lsTreeEntries                    []*LsTreeEntry
	submodulesResults                []*SubmoduleResult
	notInitializedSubmoduleFilepaths []string
}

type SubmoduleResult struct {
	*Result
}

type LsTreeEntry struct {
	Filepath string
	object.TreeEntry
}

func (r *Result) LsTree(pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := &Result{
		repository:                       r.repository,
		tree:                             r.tree,
		treeFilepath:                     r.treeFilepath,
		lsTreeEntries:                    []*LsTreeEntry{},
		submodulesResults:                []*SubmoduleResult{},
		notInitializedSubmoduleFilepaths: []string{},
	}

	for _, lsTreeEntry := range r.lsTreeEntries {
		var entryLsTreeEntries []*LsTreeEntry
		var entrySubmodulesResults []*SubmoduleResult

		var err error
		if lsTreeEntry.Filepath == "" {
			isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.Filepath)
			if isTreeMatched {
				if debugProcess() {
					logboek.Debug.LogLn("Root tree was added")
				}
				entryLsTreeEntries = append(entryLsTreeEntries, lsTreeEntry)
			} else if shouldWalkThrough {
				if debugProcess() {
					logboek.Debug.LogLn("Root tree was checking")
				}

				entryLsTreeEntries, entrySubmodulesResults, err = lsTreeWalk(r.repository, r.tree, r.treeFilepath, pathMatcher)
				if err != nil {
					return nil, err
				}
			}
		} else {
			entryLsTreeEntries, entrySubmodulesResults, err = lsTreeEntryMatch(r.repository, r.tree, r.treeFilepath, lsTreeEntry, pathMatcher)
		}

		res.lsTreeEntries = append(res.lsTreeEntries, entryLsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, entrySubmodulesResults...)
	}

	for _, submoduleResult := range r.submodulesResults {
		sr, err := submoduleResult.LsTree(pathMatcher)
		if err != nil {
			return nil, err
		}

		if !sr.IsEmpty() {
			res.submodulesResults = append(res.submodulesResults, &SubmoduleResult{sr})
		}
	}

	for _, submoduleFilepath := range r.notInitializedSubmoduleFilepaths {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFilepath)
		if isMatched || shouldGoThrough {
			res.notInitializedSubmoduleFilepaths = append(res.notInitializedSubmoduleFilepaths, submoduleFilepath)
		}
	}

	return res, nil
}

func (r *Result) Walk(f func(lsTreeEntry *LsTreeEntry) error) error {
	if err := r.lsTreeEntriesWalk(f); err != nil {
		return err
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].treeFilepath < r.submodulesResults[j].treeFilepath
	})

	for _, submoduleResult := range r.submodulesResults {
		if err := submoduleResult.Walk(f); err != nil {
			return err
		}
	}

	return nil
}

func (r *Result) Checksum() string {
	if r.IsEmpty() {
		return ""
	}

	h := sha256.New()

	_ = r.lsTreeEntriesWalk(func(lsTreeEntry *LsTreeEntry) error {
		h.Write([]byte(lsTreeEntry.Hash.String()))

		logFilepath := lsTreeEntry.Filepath
		if logFilepath == "" {
			logFilepath = "."
		}

		logboek.Debug.LogF("Entry was added: %s -> %s\n", logFilepath, lsTreeEntry.Hash.String())

		return nil
	})

	sort.Strings(r.notInitializedSubmoduleFilepaths)
	for _, submoduleFilepath := range r.notInitializedSubmoduleFilepaths {
		checksumArg := fmt.Sprintf("-%s", filepath.ToSlash(submoduleFilepath))
		h.Write([]byte(checksumArg))
		logboek.Debug.LogF("Not initialized submodule was added: %s -> %s\n", submoduleFilepath, checksumArg)
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].treeFilepath < r.submodulesResults[j].treeFilepath
	})

	for _, submoduleResult := range r.submodulesResults {
		var submoduleChecksum string
		if !submoduleResult.IsEmpty() {
			blockMsg := fmt.Sprintf("submodule %s", submoduleResult.treeFilepath)
			_ = logboek.Debug.LogBlock(blockMsg, logboek.LevelLogBlockOptions{}, func() error {
				submoduleChecksum = submoduleResult.Checksum()
				logboek.Debug.LogLn()
				logboek.Debug.LogLn(submoduleChecksum)
				return nil
			})
		}

		if submoduleChecksum != "" {
			h.Write([]byte(submoduleChecksum))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Result) IsEmpty() bool {
	return len(r.lsTreeEntries) == 0 && len(r.submodulesResults) == 0 && len(r.notInitializedSubmoduleFilepaths) == 0
}

func (r *Result) lsTreeEntriesWalk(f func(entry *LsTreeEntry) error) error {
	sort.Slice(r.lsTreeEntries, func(i, j int) bool {
		return r.lsTreeEntries[i].Filepath < r.lsTreeEntries[j].Filepath
	})

	for _, lsTreeEntry := range r.lsTreeEntries {
		if err := f(lsTreeEntry); err != nil {
			return err
		}
	}

	return nil
}
