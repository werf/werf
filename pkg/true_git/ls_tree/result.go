package ls_tree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
)

type Result struct {
	repository                           *git.Repository
	repositoryFullFilepath               string
	tree                                 *object.Tree
	lsTreeEntries                        []*LsTreeEntry
	submodulesResults                    []*SubmoduleResult
	notInitializedSubmoduleFullFilepaths []string
}

type SubmoduleResult struct {
	*Result
}

type LsTreeEntry struct {
	FullFilepath string
	object.TreeEntry
}

func (r *Result) LsTree(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := &Result{
		repository:                           r.repository,
		repositoryFullFilepath:               r.repositoryFullFilepath,
		tree:                                 r.tree,
		lsTreeEntries:                        []*LsTreeEntry{},
		submodulesResults:                    []*SubmoduleResult{},
		notInitializedSubmoduleFullFilepaths: []string{},
	}

	for _, lsTreeEntry := range r.lsTreeEntries {
		var entryLsTreeEntries []*LsTreeEntry
		var entrySubmodulesResults []*SubmoduleResult

		var err error
		if lsTreeEntry.FullFilepath == "" {
			isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.FullFilepath)
			if isTreeMatched {
				if debugProcess() {
					logboek.Context(ctx).Debug().LogLn("Root tree was added")
				}
				entryLsTreeEntries = append(entryLsTreeEntries, lsTreeEntry)
			} else if shouldWalkThrough {
				if debugProcess() {
					logboek.Context(ctx).Debug().LogLn("Root tree was checking")
				}

				entryLsTreeEntries, entrySubmodulesResults, err = lsTreeWalk(ctx, r.repository, r.tree, r.repositoryFullFilepath, r.repositoryFullFilepath, pathMatcher)
				if err != nil {
					return nil, err
				}
			}
		} else {
			entryLsTreeEntries, entrySubmodulesResults, err = lsTreeEntryMatch(ctx, r.repository, r.tree, r.repositoryFullFilepath, r.repositoryFullFilepath, lsTreeEntry, pathMatcher)
			if err != nil {
				return nil, err
			}
		}

		res.lsTreeEntries = append(res.lsTreeEntries, entryLsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, entrySubmodulesResults...)
	}

	for _, submoduleResult := range r.submodulesResults {
		sr, err := submoduleResult.LsTree(ctx, pathMatcher)
		if err != nil {
			return nil, err
		}

		if !sr.IsEmpty() {
			res.submodulesResults = append(res.submodulesResults, &SubmoduleResult{sr})
		}
	}

	for _, submoduleFullFilepath := range r.notInitializedSubmoduleFullFilepaths {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFullFilepath)
		if isMatched || shouldGoThrough {
			res.notInitializedSubmoduleFullFilepaths = append(res.notInitializedSubmoduleFullFilepaths, submoduleFullFilepath)
		}
	}

	return res, nil
}

func (r *Result) Walk(f func(lsTreeEntry *LsTreeEntry) error) error {
	if err := r.lsTreeEntriesWalk(f); err != nil {
		return err
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].repositoryFullFilepath < r.submodulesResults[j].repositoryFullFilepath
	})

	for _, submoduleResult := range r.submodulesResults {
		if err := submoduleResult.Walk(f); err != nil {
			return err
		}
	}

	return nil
}

func (r *Result) Checksum(ctx context.Context) string {
	if r.IsEmpty() {
		return ""
	}

	h := sha256.New()

	_ = r.lsTreeEntriesWalk(func(lsTreeEntry *LsTreeEntry) error {
		h.Write([]byte(lsTreeEntry.Hash.String()))

		logFilepath := lsTreeEntry.FullFilepath
		if logFilepath == "" {
			logFilepath = "."
		}

		logboek.Context(ctx).Debug().LogF("Entry was added: %s -> %s\n", logFilepath, lsTreeEntry.Hash.String())

		return nil
	})

	sort.Strings(r.notInitializedSubmoduleFullFilepaths)
	for _, submoduleFullFilepath := range r.notInitializedSubmoduleFullFilepaths {
		checksumArg := fmt.Sprintf("-%s", filepath.ToSlash(submoduleFullFilepath))
		h.Write([]byte(checksumArg))
		logboek.Context(ctx).Debug().LogF("Not initialized submodule was added: %s -> %s\n", submoduleFullFilepath, checksumArg)
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].repositoryFullFilepath < r.submodulesResults[j].repositoryFullFilepath
	})

	for _, submoduleResult := range r.submodulesResults {
		var submoduleChecksum string
		if !submoduleResult.IsEmpty() {
			logboek.Context(ctx).Debug().LogOptionalLn()
			blockMsg := fmt.Sprintf("submodule %s", submoduleResult.repositoryFullFilepath)
			logboek.Context(ctx).Debug().LogBlock(blockMsg).Do(func() {
				submoduleChecksum = submoduleResult.Checksum(ctx)
				logboek.Context(ctx).Debug().LogLn()
				logboek.Context(ctx).Debug().LogLn(submoduleChecksum)
			})
		}

		if submoduleChecksum != "" {
			h.Write([]byte(submoduleChecksum))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Result) IsEmpty() bool {
	return len(r.lsTreeEntries) == 0 && len(r.submodulesResults) == 0 && len(r.notInitializedSubmoduleFullFilepaths) == 0
}

func (r *Result) lsTreeEntriesWalk(f func(entry *LsTreeEntry) error) error {
	sort.Slice(r.lsTreeEntries, func(i, j int) bool {
		return r.lsTreeEntries[i].FullFilepath < r.lsTreeEntries[j].FullFilepath
	})

	for _, lsTreeEntry := range r.lsTreeEntries {
		if err := f(lsTreeEntry); err != nil {
			return err
		}
	}

	return nil
}
