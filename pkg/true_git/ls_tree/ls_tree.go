package ls_tree

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/path_matcher"
)

func newHash(s string) (plumbing.Hash, error) {
	var h plumbing.Hash

	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}

	copy(h[:], b)
	return h, nil
}

// LsTree returns the Result with tree entries that satisfy the passed pathMatcher.
// The function works lazily and does not go through a tree directory unnecessarily.
// If the result should contain only regular files (without directories and submodules), you should use the allFiles parameter.
func LsTree(ctx context.Context, repository *git.Repository, commit string, pathMatcher path_matcher.PathMatcher, allFiles bool) (*Result, error) {
	r, err := lsTree(ctx, repository, commit, pathMatcher, allFiles)
	if err != nil {
		return nil, err
	}

	r.setParentRecursively()
	return r, nil
}

func lsTree(ctx context.Context, repository *git.Repository, commit string, pathMatcher path_matcher.PathMatcher, allFiles bool) (*Result, error) {
	res := NewResult(commit, "", []*LsTreeEntry{}, []*SubmoduleResult{})

	tree, err := getCommitTree(repository, commit)
	if err != nil {
		return nil, err
	}

	baseFilepath := pathMatcher.BaseFilepath()
	if baseFilepath != "" {
		lsTreeEntries, submodulesResults, err := processSpecificEntryFilepath(ctx, repository, tree, "", "", pathMatcher.BaseFilepath(), pathMatcher, allFiles)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = append(res.lsTreeEntries, lsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, submodulesResults...)

		return res, nil
	}

	rootEntry := ""
	if err := lsTreeDirOrSubmoduleEntryMatchBase(
		rootEntry,
		pathMatcher,
		allFiles,
		// add tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Root tree was added")
			}

			rootTreeEntry := &LsTreeEntry{
				FullFilepath: "",
				TreeEntry: object.TreeEntry{
					Name: "",
					Mode: filemode.Dir,
					Hash: tree.Hash,
				},
			}

			res.lsTreeEntries = append(res.lsTreeEntries, rootTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Root tree was checking")
			}

			lsTreeEntries, submodulesLsTreeEntries, err := lsTreeWalk(ctx, repository, tree, "", "", pathMatcher, allFiles)
			if err != nil {
				return err
			}

			res.lsTreeEntries = lsTreeEntries
			res.submodulesResults = submodulesLsTreeEntries

			return nil
		},
		// skip tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Root tree was skipped")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return res, nil
}

func getCommitTree(repository *git.Repository, commit string) (*object.Tree, error) {
	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("invalid commit %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s commit info: %s", commit, err)
	}

	tree, err := commitObj.Tree()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func processSpecificEntryFilepath(ctx context.Context, repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath, treeEntryFilepath string, pathMatcher path_matcher.PathMatcher, allFiles bool) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, nil, err
	}

	submodules, err := worktree.Submodules()
	for _, submodule := range submodules {
		submoduleEntryFilepath := filepath.FromSlash(submodule.Config().Path)
		submoduleFullFilepath := filepath.Join(treeFullFilepath, submoduleEntryFilepath)
		relTreeEntryFilepath, err := filepath.Rel(submoduleEntryFilepath, treeEntryFilepath)
		if err != nil {
			panic(err)
		}

		if relTreeEntryFilepath == "." || relTreeEntryFilepath == ".." || strings.HasPrefix(relTreeEntryFilepath, ".."+string(os.PathSeparator)) {
			continue
		}

		submoduleRepository, submoduleExpectedCommit, submoduleTree, err := submoduleRepositoryAndTree(ctx, repository, submodule.Config().Path)
		if err != nil {
			return nil, nil, fmt.Errorf("getting submodule %q repository and tree failed: %s", submoduleFullFilepath, err)
		}

		sLsTreeEntries, sSubmodulesResults, err := processSpecificEntryFilepath(ctx, submoduleRepository, submoduleTree, submoduleFullFilepath, submoduleFullFilepath, relTreeEntryFilepath, pathMatcher, allFiles)
		if err != nil {
			return nil, nil, err
		}

		result := NewResult(submoduleExpectedCommit, submoduleFullFilepath, sLsTreeEntries, sSubmodulesResults)
		submoduleResult := NewSubmoduleResult(submodule.Config().Name, submodule.Config().Path, result)

		if !submoduleResult.IsEmpty() {
			submodulesResults = append(submodulesResults, submoduleResult)
		}

		return lsTreeEntries, submodulesResults, nil
	}

	lsTreeEntry, err := treeFindEntry(ctx, tree, treeFullFilepath, treeEntryFilepath)
	if err != nil {
		if err == object.ErrDirectoryNotFound || err == object.ErrFileNotFound || err == object.ErrEntryNotFound || err == plumbing.ErrObjectNotFound {
			return lsTreeEntries, submodulesResults, nil
		}

		return nil, nil, err
	}

	lsTreeEntries, submodulesLsTreeEntries, err := lsTreeEntryMatch(ctx, repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher, allFiles)
	if err != nil {
		return nil, nil, err
	}

	return lsTreeEntries, submodulesLsTreeEntries, nil
}

func lsTreeWalk(ctx context.Context, repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, pathMatcher path_matcher.PathMatcher, allFiles bool) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	for _, treeEntry := range tree.Entries {
		lsTreeEntry := &LsTreeEntry{
			FullFilepath: filepath.Join(treeFullFilepath, treeEntry.Name),
			TreeEntry:    treeEntry,
		}

		entryTreeEntries, entrySubmodulesTreeEntries, err := lsTreeEntryMatch(ctx, repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher, allFiles)
		if err != nil {
			return nil, nil, err
		}

		lsTreeEntries = append(lsTreeEntries, entryTreeEntries...)
		submodulesResults = append(submodulesResults, entrySubmodulesTreeEntries...)
	}

	return
}

func lsTreeEntryMatch(ctx context.Context, repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher, allFiles bool) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	switch lsTreeEntry.Mode {
	case filemode.Dir:
		return lsTreeDirEntryMatch(ctx, repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher, allFiles)
	case filemode.Submodule:
		return lsTreeSubmoduleEntryMatch(ctx, repository, repositoryFullFilepath, lsTreeEntry, pathMatcher, allFiles)
	default:
		return lsTreeFileEntryMatch(ctx, lsTreeEntry, pathMatcher)
	}
}

func lsTreeDirEntryMatch(ctx context.Context, repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher, allFiles bool) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if err := lsTreeDirOrSubmoduleEntryMatchBase(
		lsTreeEntry.FullFilepath,
		pathMatcher,
		allFiles,
		// add tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was added:         ", lsTreeEntry.FullFilepath)
			}

			lsTreeEntries = append(lsTreeEntries, lsTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was checking:      ", lsTreeEntry.FullFilepath)
			}

			entryTree, err := treeTree(tree, treeFullFilepath, lsTreeEntry.FullFilepath)
			if err != nil {
				return err
			}

			lsTreeEntries, submodulesResults, err = lsTreeWalk(ctx, repository, entryTree, repositoryFullFilepath, lsTreeEntry.FullFilepath, pathMatcher, allFiles)
			if err != nil {
				return err
			}

			return nil
		},
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was skipped:       ", lsTreeEntry.FullFilepath)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return
}

func lsTreeSubmoduleEntryMatch(ctx context.Context, repository *git.Repository, repositoryFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher, allFiles bool) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if err := lsTreeDirOrSubmoduleEntryMatchBase(
		lsTreeEntry.FullFilepath,
		pathMatcher,
		allFiles,
		// add tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Submodule entry was added:   ", lsTreeEntry.FullFilepath)
			}
			lsTreeEntries = append(lsTreeEntries, lsTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debugProcess() {
				logboek.Context(ctx).Debug().LogLn("Submodule entry was checking:", lsTreeEntry.FullFilepath)
			}

			submoduleFilepath, err := filepath.Rel(repositoryFullFilepath, lsTreeEntry.FullFilepath)
			if err != nil || submoduleFilepath == "." || submoduleFilepath == ".." || strings.HasPrefix(submoduleFilepath, ".."+string(os.PathSeparator)) {
				panic(fmt.Sprintf("unexpected paths: %s, %s", repositoryFullFilepath, lsTreeEntry.FullFilepath))
			}

			submodulePath := filepath.ToSlash(submoduleFilepath)
			submoduleRepository, submoduleCommit, submoduleTree, err := submoduleRepositoryAndTree(ctx, repository, submodulePath)
			if err != nil {
				return fmt.Errorf("getting submodule %q repository and tree failed: %s", lsTreeEntry.FullFilepath, err)
			}

			submoduleLsTreeEntrees, submoduleSubmoduleResults, err := lsTreeWalk(ctx, submoduleRepository, submoduleTree, lsTreeEntry.FullFilepath, lsTreeEntry.FullFilepath, pathMatcher, allFiles)
			if err != nil {
				return err
			}

			if len(submoduleLsTreeEntrees) != 0 {
				w, err := repository.Worktree()
				if err != nil {
					return err
				}

				ss, err := w.Submodules()
				if err != nil {
					return err
				}

				var submoduleName string
				for _, s := range ss {
					if s.Config().Path == submodulePath {
						submoduleName = s.Config().Name
					}
				}

				if submoduleName == "" {
					panic("unexpected condition " + submodulePath)
				}

				result := NewResult(submoduleCommit, lsTreeEntry.FullFilepath, submoduleLsTreeEntrees, submoduleSubmoduleResults)
				submoduleResult := NewSubmoduleResult(submoduleName, submodulePath, result)

				if !submoduleResult.IsEmpty() {
					submodulesResults = append(submodulesResults, submoduleResult)
				}
			}

			return nil
		},
		// skip tree func
		func() error {
			logboek.Context(ctx).Debug().LogLn("Submodule entry was skipped: ", lsTreeEntry.FullFilepath)
			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return
}

func lsTreeDirOrSubmoduleEntryMatchBase(path string, pathMatched path_matcher.PathMatcher, allFiles bool, addTreeFunc, checkTreeFunc, skipTreeFunc func() error) error {
	if pathMatched.ShouldGoThrough(path) {
		return checkTreeFunc()
	} else if pathMatched.IsPathMatched(path) {
		if allFiles {
			return checkTreeFunc()
		} else {
			return addTreeFunc()
		}
	} else {
		return skipTreeFunc()
	}
}

func lsTreeFileEntryMatch(ctx context.Context, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if pathMatcher.IsPathMatched(lsTreeEntry.FullFilepath) {
		if debugProcess() {
			logboek.Context(ctx).Debug().LogLn("File entry was added:        ", lsTreeEntry.FullFilepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	}

	return
}

func treeFindEntry(_ context.Context, tree *object.Tree, treeFullFilepath, treeEntryFilepath string) (*LsTreeEntry, error) {
	formattedTreeEntryPath := filepath.ToSlash(treeEntryFilepath)
	treeEntry, err := tree.FindEntry(formattedTreeEntryPath)
	if err != nil {
		return nil, err
	}

	return &LsTreeEntry{
		FullFilepath: filepath.Join(treeFullFilepath, treeEntryFilepath),
		TreeEntry:    *treeEntry,
	}, nil
}

func treeTree(tree *object.Tree, treeFullFilepath, treeDirEntryFullFilepath string) (*object.Tree, error) {
	treeDirEntryFilepath, err := filepath.Rel(treeFullFilepath, treeDirEntryFullFilepath)
	if err != nil || treeDirEntryFilepath == "." || treeDirEntryFilepath == ".." || strings.HasPrefix(treeDirEntryFilepath, ".."+string(os.PathSeparator)) {
		panic(fmt.Sprintf("unexpected paths: %s, %s", treeFullFilepath, treeDirEntryFullFilepath))
	}

	treeDirEntryPath := filepath.ToSlash(treeDirEntryFilepath)
	entryTree, err := tree.Tree(treeDirEntryPath)
	if err != nil {
		return nil, err
	}

	return entryTree, nil
}

func submoduleRepositoryAndTree(ctx context.Context, repository *git.Repository, submodulePath string) (*git.Repository, string, *object.Tree, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot inspect worktree: %s", err)
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot get repository submodules: %s", err)
	}

	var submodule *git.Submodule
	for _, s := range submodules {
		if s.Config().Path == submodulePath {
			submodule = s
			break
		}
	}

	if submodule == nil {
		return nil, "", nil, fmt.Errorf("cannot get submodule by path %s", submodulePath)
	}

	submoduleRepository, err := submodule.Repository()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot inspect submodule %q repository: %s", submodulePath, err)
	}

	submoduleStatus, err := submodule.Status()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot get submodule %q status: %s", submodulePath, err)
	}

	if debugProcess() {
		if !submoduleStatus.IsClean() {
			logboek.Context(ctx).Debug().LogFWithCustomStyle(
				color.GetStyle("danger"),
				"Submodule is not clean (current commit %s), expected commit %s will be checked\n",
				submoduleStatus.Current,
				submoduleStatus.Expected,
			)
		}
	}

	commit, err := submoduleRepository.CommitObject(submoduleStatus.Expected)
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot inspect submodule %q commit %q: %s", submodulePath, submoduleStatus.Expected, err)
	}

	submoduleTree, err := commit.Tree()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot inspect submodule %q commit %q tree: %s", submodulePath, submoduleStatus.Expected, err)
	}

	return submoduleRepository, submoduleStatus.Expected.String(), submoduleTree, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_LS_TREE_PROCESS") == "1"
}
