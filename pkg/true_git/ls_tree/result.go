package ls_tree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/git_repo/repo_handle"
)

type Result struct {
	commit                 string
	repositoryFullFilepath string
	lsTreeEntries          []*LsTreeEntry
	submodulesResults      []*SubmoduleResult
}

func NewResult(commit, repositoryFullFilepath string, lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult) *Result {
	return &Result{
		commit:                 commit,
		repositoryFullFilepath: repositoryFullFilepath,
		lsTreeEntries:          lsTreeEntries,
		submodulesResults:      submodulesResults,
	}
}

func (r *Result) setParentRecursively() {
	for _, s := range r.submodulesResults {
		s.parentResult = r
		s.setParentRecursively()
	}
}

func NewSubmoduleResult(submoduleName, submodulePath string, result *Result) *SubmoduleResult {
	return &SubmoduleResult{
		submoduleName: submoduleName,
		submodulePath: submodulePath,
		Result:        result,
	}
}

type SubmoduleResult struct {
	*Result
	submoduleName string
	submodulePath string

	parentResult          *Result
	parentSubmoduleResult *SubmoduleResult
}

func (r *SubmoduleResult) setParentRecursively() {
	for _, s := range r.submodulesResults {
		s.parentSubmoduleResult = r
		s.setParentRecursively()
	}
}

func (r *SubmoduleResult) submoduleRepositoryFromParent(mainRepository repo_handle.Handle) (repo_handle.Handle, error) {
	var parentRepository repo_handle.Handle
	switch {
	case r.parentResult != nil:
		parentRepository = mainRepository
	case r.parentSubmoduleResult != nil:
		var err error
		parentRepository, err = r.parentSubmoduleResult.submoduleRepositoryFromParent(mainRepository)
		if err != nil {
			return nil, err
		}
	default:
		panic("unexpected condition")
	}

	return r.submoduleRepository(parentRepository)
}

func (r *SubmoduleResult) submoduleRepository(parentRepoHandle repo_handle.Handle) (repo_handle.Handle, error) {
	return parentRepoHandle.Submodule(r.submodulePath)
}

type LsTreeEntry struct {
	FullFilepath string
	object.TreeEntry
}

func (r *Result) LsTree(ctx context.Context, repoHandle repo_handle.Handle, opts LsTreeOptions) (*Result, error) {
	r, err := r.lsTree(ctx, repoHandle, opts)
	if err != nil {
		return nil, err
	}

	r.setParentRecursively()
	return r, nil
}

func (r *Result) lsTree(ctx context.Context, repoHandle repo_handle.Handle, opts LsTreeOptions) (*Result, error) {
	res := NewResult(r.commit, r.repositoryFullFilepath, []*LsTreeEntry{}, []*SubmoduleResult{})

	tree, err := repoHandle.GetCommitTree(plumbing.NewHash(r.commit))
	if err != nil {
		return nil, err
	}

	for _, lsTreeEntry := range r.lsTreeEntries {
		var entryLsTreeEntries []*LsTreeEntry
		var entrySubmodulesResults []*SubmoduleResult

		var err error
		if lsTreeEntry.FullFilepath == "" {
			if err := lsTreeDirOrSubmoduleEntryMatchBase(
				lsTreeEntry.FullFilepath,
				opts,
				// add tree func
				func() error {
					if debug() {
						logboek.Context(ctx).Debug().LogLn("Root tree was added")
					}

					entryLsTreeEntries = append(entryLsTreeEntries, lsTreeEntry)

					return nil
				},
				// check tree func
				func() error {
					if debug() {
						logboek.Context(ctx).Debug().LogLn("Root tree was checking")
					}

					entryLsTreeEntries, entrySubmodulesResults, err = lsTreeWalk(ctx, repoHandle, tree, r.repositoryFullFilepath, r.repositoryFullFilepath, opts)
					return err
				},
				// skip tree func
				func() error {
					if debug() {
						logboek.Context(ctx).Debug().LogLn("Root tree was skipped")
					}

					return nil
				},
			); err != nil {
				return nil, err
			}
		} else {
			entryLsTreeEntries, entrySubmodulesResults, err = lsTreeEntryMatch(ctx, repoHandle, tree, r.repositoryFullFilepath, r.repositoryFullFilepath, lsTreeEntry, opts)
			if err != nil {
				return nil, err
			}
		}

		res.lsTreeEntries = append(res.lsTreeEntries, entryLsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, entrySubmodulesResults...)
	}

	for _, submoduleResult := range r.submodulesResults {
		submoduleRepository, err := submoduleResult.submoduleRepository(repoHandle)
		if err != nil {
			return nil, err
		}

		sr, err := submoduleResult.lsTree(ctx, submoduleRepository, opts)
		if err != nil {
			return nil, err
		}

		if !sr.IsEmpty() {
			sResult := NewSubmoduleResult(submoduleResult.submoduleName, submoduleResult.submodulePath, sr)
			res.submodulesResults = append(res.submodulesResults, sResult)
		}
	}

	return res, nil
}

func (r *Result) Walk(f func(lsTreeEntry *LsTreeEntry) error) error {
	return r.walkWithResult(func(_ *Result, _ *SubmoduleResult, lsTreeEntry *LsTreeEntry) error {
		return f(lsTreeEntry)
	})
}

func (r *Result) walkWithResult(f func(r *Result, sr *SubmoduleResult, lsTreeEntry *LsTreeEntry) error) error {
	if err := r.lsTreeEntriesWalkWithResult(f); err != nil {
		return err
	}

	sort.Slice(r.submodulesResults, func(i, j int) bool {
		return r.submodulesResults[i].repositoryFullFilepath < r.submodulesResults[j].repositoryFullFilepath
	})

	for _, submoduleResult := range r.submodulesResults {
		if err := submoduleResult.walkWithResult(func(_ *Result, _ *SubmoduleResult, lsTreeEntry *LsTreeEntry) error {
			return f(nil, submoduleResult, lsTreeEntry)
		}); err != nil {
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
	return len(r.lsTreeEntries) == 0 && len(r.submodulesResults) == 0
}

func (r *Result) lsTreeEntriesWalk(f func(entry *LsTreeEntry) error) error {
	return r.lsTreeEntriesWalkWithResult(func(_ *Result, _ *SubmoduleResult, entry *LsTreeEntry) error {
		return f(entry)
	})
}

func (r *Result) lsTreeEntriesWalkWithResult(f func(r *Result, sr *SubmoduleResult, lsTreeEntry *LsTreeEntry) error) error {
	sort.Slice(r.lsTreeEntries, func(i, j int) bool {
		return r.lsTreeEntries[i].FullFilepath < r.lsTreeEntries[j].FullFilepath
	})

	for _, lsTreeEntry := range r.lsTreeEntries {
		if err := f(r, nil, lsTreeEntry); err != nil {
			return err
		}
	}

	return nil
}

func (r *Result) LsTreeEntry(relPath string) *LsTreeEntry {
	if relPath == "." {
		relPath = ""
	}

	var lsTreeEntry *LsTreeEntry
	_ = r.Walk(func(entry *LsTreeEntry) error {
		if filepath.ToSlash(entry.FullFilepath) == filepath.ToSlash(relPath) {
			lsTreeEntry = entry
		}

		return nil
	})

	if lsTreeEntry == nil {
		lsTreeEntry = &LsTreeEntry{
			FullFilepath: relPath,
			TreeEntry: object.TreeEntry{
				Name: relPath,
				Mode: filemode.Empty,
				Hash: plumbing.Hash{},
			},
		}
	}

	return lsTreeEntry
}

func (r *Result) LsTreeEntryContent(mainRepoHandle repo_handle.Handle, relPath string) ([]byte, error) {
	var entryRepoHandle repo_handle.Handle
	var entry *LsTreeEntry

	_ = r.walkWithResult(func(r *Result, sr *SubmoduleResult, e *LsTreeEntry) (err error) {
		if filepath.ToSlash(e.FullFilepath) == filepath.ToSlash(relPath) {
			switch {
			case r != nil:
				entryRepoHandle = mainRepoHandle
			case sr != nil:
				entryRepoHandle, err = sr.submoduleRepositoryFromParent(mainRepoHandle)
				if err != nil {
					return err
				}
			default:
				panic(fmt.Sprintf("unexpected condition: %v", e))
			}

			entry = e
		}

		return nil
	})

	if entry == nil {
		return nil, fmt.Errorf("unable to get tree entry %s", relPath)
	}

	if entryRepoHandle == nil {
		panic("unexpected condition")
	}

	return entryRepoHandle.ReadBlobObjectContent(entry.Hash)
}
