package ls_tree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
)

type Result struct {
	commit                                  string
	repositoryFullFilepath                  string
	lsTreeEntries                           []*LsTreeEntry
	submodulesResults                       []*SubmoduleResult
	notInitializedSubmoduleFullFilepathList []string
}

func NewResult(commit string, repositoryFullFilepath string, lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, notInitializedSubmoduleFullFilepathList []string) *Result {
	return &Result{
		commit:                                  commit,
		repositoryFullFilepath:                  repositoryFullFilepath,
		lsTreeEntries:                           lsTreeEntries,
		submodulesResults:                       submodulesResults,
		notInitializedSubmoduleFullFilepathList: notInitializedSubmoduleFullFilepathList,
	}
}

func (r *Result) setParentRecursively() {
	for _, s := range r.submodulesResults {
		s.parentResult = r
		s.setParentRecursively()
	}
}

func NewSubmoduleResult(submodulePath, submoduleName string, result *Result) *SubmoduleResult {
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

func (r *SubmoduleResult) submoduleRepositoryFromParent(mainRepository *git.Repository) (*git.Repository, error) {
	var parentRepository *git.Repository
	if r.parentResult != nil {
		parentRepository = mainRepository
	} else if r.parentSubmoduleResult != nil {
		var err error
		parentRepository, err = r.parentSubmoduleResult.submoduleRepositoryFromParent(mainRepository)
		if err != nil {
			return nil, err
		}
	} else {
		panic("unexpected condition")
	}

	return r.submoduleRepository(parentRepository)
}

func (r *SubmoduleResult) submoduleRepository(parentRepository *git.Repository) (*git.Repository, error) {
	w, err := parentRepository.Worktree()
	if err != nil {
		return nil, err
	}

	s, err := w.Submodule(r.submoduleName)
	if err != nil {
		return nil, err
	}

	submoduleRepository, err := s.Repository()
	if err != nil {
		return nil, err
	}

	return submoduleRepository, nil
}

type LsTreeEntry struct {
	FullFilepath string
	object.TreeEntry
}

func (r *Result) LsTree(ctx context.Context, repository *git.Repository, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	r, err := r.lsTree(ctx, repository, pathMatcher)
	if err != nil {
		return nil, err
	}

	r.setParentRecursively()
	return r, nil
}

func (r *Result) lsTree(ctx context.Context, repository *git.Repository, pathMatcher path_matcher.PathMatcher) (*Result, error) {
	res := NewResult(r.commit, r.repositoryFullFilepath, []*LsTreeEntry{}, []*SubmoduleResult{}, []string{})

	tree, err := getCommitTree(repository, r.commit)
	if err != nil {
		return nil, err
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

				entryLsTreeEntries, entrySubmodulesResults, err = lsTreeWalk(ctx, repository, tree, r.repositoryFullFilepath, r.repositoryFullFilepath, pathMatcher)
				if err != nil {
					return nil, err
				}
			}
		} else {
			entryLsTreeEntries, entrySubmodulesResults, err = lsTreeEntryMatch(ctx, repository, tree, r.repositoryFullFilepath, r.repositoryFullFilepath, lsTreeEntry, pathMatcher)
			if err != nil {
				return nil, err
			}
		}

		res.lsTreeEntries = append(res.lsTreeEntries, entryLsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, entrySubmodulesResults...)
	}

	for _, submoduleResult := range r.submodulesResults {
		submoduleRepository, err := submoduleResult.submoduleRepository(repository)
		if err != nil {
			return nil, err
		}

		sr, err := submoduleResult.lsTree(ctx, submoduleRepository, pathMatcher)
		if err != nil {
			return nil, err
		}

		if !sr.IsEmpty() {
			sResult := NewSubmoduleResult(submoduleResult.submoduleName, submoduleResult.submodulePath, sr)
			res.submodulesResults = append(res.submodulesResults, sResult)
		}
	}

	for _, submoduleFullFilepath := range r.notInitializedSubmoduleFullFilepathList {
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFullFilepath)
		if isMatched || shouldGoThrough {
			res.notInitializedSubmoduleFullFilepathList = append(res.notInitializedSubmoduleFullFilepathList, submoduleFullFilepath)
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

	sort.Strings(r.notInitializedSubmoduleFullFilepathList)
	for _, submoduleFullFilepath := range r.notInitializedSubmoduleFullFilepathList {
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
	return len(r.lsTreeEntries) == 0 && len(r.submodulesResults) == 0 && len(r.notInitializedSubmoduleFullFilepathList) == 0
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

func (r *Result) LsTreeEntryContent(mainRepository *git.Repository, relPath string) ([]byte, error) {
	var entryRepository *git.Repository
	var entry *LsTreeEntry

	_ = r.walkWithResult(func(r *Result, sr *SubmoduleResult, e *LsTreeEntry) (err error) {
		if filepath.ToSlash(e.FullFilepath) == filepath.ToSlash(relPath) {
			if r != nil {
				entryRepository = mainRepository
			} else if sr != nil {
				entryRepository, err = sr.submoduleRepositoryFromParent(mainRepository)
				if err != nil {
					return err
				}
			} else {
				panic(fmt.Sprintf("unexpected condition: %v", e))
			}
			entry = e
		}

		return nil
	})

	if entry == nil {
		return nil, fmt.Errorf("unable to get tree entry %s", relPath)
	}

	if entryRepository == nil {
		panic("unexpected condition")
	}

	obj, err := entryRepository.BlobObject(entry.Hash)
	if err != nil {
		return nil, fmt.Errorf("unable to get tree entry %q blob object: %s", entry.FullFilepath, err)
	}

	f, err := obj.Reader()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read tree entry %q content: %s", relPath, err)
	}

	return data, err
}
