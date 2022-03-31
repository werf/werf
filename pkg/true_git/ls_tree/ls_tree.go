package ls_tree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

type LsTreeOptions struct {
	// the PathScope option determines the directory or file that will get into the result (similar to <pathspec> in the git commands)
	PathScope   string
	PathMatcher path_matcher.PathMatcher
	AllFiles    bool
}

func (opts LsTreeOptions) ID() string {
	return util.Sha256Hash(
		opts.PathScope,
		opts.PathMatcher.ID(),
		fmt.Sprint(opts.AllFiles),
	)
}

func (opts LsTreeOptions) formattedPathScope() string {
	if opts.PathScope == "." {
		return ""
	}

	return filepath.FromSlash(opts.PathScope)
}

func (opts LsTreeOptions) formattedPathMatcher() path_matcher.PathMatcher {
	var matchers []path_matcher.PathMatcher
	matchers = append(matchers, path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: opts.formattedPathScope()}))
	if opts.PathMatcher != nil {
		matchers = append(matchers, opts.PathMatcher)
	}

	return path_matcher.NewMultiPathMatcher(matchers...)
}

// LsTree returns the Result with tree entries that satisfy the passed options.
// The function works lazily and does not go through a tree directory unnecessarily.
// If the result should contain only regular files (without directories and submodules), you should use the AllFiles option.
func LsTree(ctx context.Context, repoHandle repo_handle.Handle, commit string, opts LsTreeOptions) (*Result, error) {
	r, err := lsTree(ctx, repoHandle, commit, opts)
	if err != nil {
		return nil, err
	}

	r.setParentRecursively()
	return r, nil
}

func lsTree(ctx context.Context, repoHandle repo_handle.Handle, commit string, opts LsTreeOptions) (*Result, error) {
	res := NewResult(commit, "", []*LsTreeEntry{}, []*SubmoduleResult{})

	tree, err := repoHandle.GetCommitTree(plumbing.NewHash(commit))
	if err != nil {
		return nil, err
	}

	if opts.formattedPathScope() != "" {
		lsTreeEntries, submodulesResults, err := processSpecificEntryFilepath(ctx, repoHandle, tree, "", "", opts.formattedPathScope(), opts)
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
		opts,
		// add tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Root tree was added")
			}

			rootTreeEntry := &LsTreeEntry{
				FullFilepath: "",
				TreeEntry: object.TreeEntry{
					Name: "",
					Mode: filemode.Dir,
					Hash: tree.Hash(),
				},
			}

			res.lsTreeEntries = append(res.lsTreeEntries, rootTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Root tree was checking")
			}

			lsTreeEntries, submodulesLsTreeEntries, err := lsTreeWalk(ctx, repoHandle, tree, "", "", opts)
			if err != nil {
				return err
			}

			res.lsTreeEntries = lsTreeEntries
			res.submodulesResults = submodulesLsTreeEntries

			return nil
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

	return res, nil
}

func processSpecificEntryFilepath(ctx context.Context, repoHandle repo_handle.Handle, tree repo_handle.TreeHandle, repositoryFullFilepath, treeFullFilepath, treeEntryFilepath string, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	for _, submoduleHandle := range repoHandle.Submodules() {
		submoduleEntryFilepath := filepath.FromSlash(submoduleHandle.Config().Path)
		submoduleFullFilepath := filepath.Join(treeFullFilepath, submoduleEntryFilepath)
		relTreeEntryFilepath, err := filepath.Rel(submoduleEntryFilepath, treeEntryFilepath)
		if err != nil {
			panic(err)
		}

		if relTreeEntryFilepath == "." || relTreeEntryFilepath == ".." || strings.HasPrefix(relTreeEntryFilepath, ".."+string(os.PathSeparator)) {
			continue
		}

		submoduleTree, err := getSubmoduleTree(submoduleHandle)
		if err != nil {
			return nil, nil, err
		}

		sLsTreeEntries, sSubmodulesResults, err := processSpecificEntryFilepath(ctx, submoduleHandle, submoduleTree, submoduleFullFilepath, submoduleFullFilepath, relTreeEntryFilepath, opts)
		if err != nil {
			return nil, nil, err
		}

		result := NewResult(submoduleHandle.Status().Expected.String(), submoduleFullFilepath, sLsTreeEntries, sSubmodulesResults)
		submoduleResult := NewSubmoduleResult(submoduleHandle.Config().Name, submoduleHandle.Config().Path, result)

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

	lsTreeEntries, submodulesLsTreeEntries, err := lsTreeEntryMatch(ctx, repoHandle, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, opts)
	if err != nil {
		return nil, nil, err
	}

	return lsTreeEntries, submodulesLsTreeEntries, nil
}

func lsTreeWalk(ctx context.Context, repoHandle repo_handle.Handle, tree repo_handle.TreeHandle, repositoryFullFilepath, treeFullFilepath string, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	for _, treeEntry := range tree.Entries() {
		lsTreeEntry := &LsTreeEntry{
			FullFilepath: filepath.Join(treeFullFilepath, treeEntry.Name),
			TreeEntry:    treeEntry,
		}

		entryTreeEntries, entrySubmodulesTreeEntries, err := lsTreeEntryMatch(ctx, repoHandle, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, opts)
		if err != nil {
			return nil, nil, err
		}

		lsTreeEntries = append(lsTreeEntries, entryTreeEntries...)
		submodulesResults = append(submodulesResults, entrySubmodulesTreeEntries...)
	}

	return
}

func lsTreeEntryMatch(ctx context.Context, repoHandle repo_handle.Handle, tree repo_handle.TreeHandle, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	switch lsTreeEntry.Mode {
	case filemode.Dir:
		return lsTreeDirEntryMatch(ctx, repoHandle, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, opts)
	case filemode.Submodule:
		return lsTreeSubmoduleEntryMatch(ctx, repoHandle, repositoryFullFilepath, lsTreeEntry, opts)
	default:
		return lsTreeFileEntryMatch(ctx, lsTreeEntry, opts)
	}
}

func lsTreeDirEntryMatch(ctx context.Context, repoHandle repo_handle.Handle, tree repo_handle.TreeHandle, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if err := lsTreeDirOrSubmoduleEntryMatchBase(
		lsTreeEntry.FullFilepath,
		opts,
		// add tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was added:         ", lsTreeEntry.FullFilepath)
			}

			lsTreeEntries = append(lsTreeEntries, lsTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was checking:      ", lsTreeEntry.FullFilepath)
			}

			entryTree, err := treeTree(tree, treeFullFilepath, lsTreeEntry.FullFilepath)
			if err != nil {
				return err
			}

			lsTreeEntries, submodulesResults, err = lsTreeWalk(ctx, repoHandle, entryTree, repositoryFullFilepath, lsTreeEntry.FullFilepath, opts)
			if err != nil {
				return err
			}

			return nil
		},
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Dir entry was skipped:       ", lsTreeEntry.FullFilepath)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return
}

func lsTreeSubmoduleEntryMatch(ctx context.Context, repoHandle repo_handle.Handle, repositoryFullFilepath string, lsTreeEntry *LsTreeEntry, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if err := lsTreeDirOrSubmoduleEntryMatchBase(
		lsTreeEntry.FullFilepath,
		opts,
		// add tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Submodule entry was added:   ", lsTreeEntry.FullFilepath)
			}
			lsTreeEntries = append(lsTreeEntries, lsTreeEntry)

			return nil
		},
		// check tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Submodule entry was checking:", lsTreeEntry.FullFilepath)
			}

			submoduleFilepath, err := filepath.Rel(repositoryFullFilepath, lsTreeEntry.FullFilepath)
			if err != nil || submoduleFilepath == "." || submoduleFilepath == ".." || strings.HasPrefix(submoduleFilepath, ".."+string(os.PathSeparator)) {
				panic(fmt.Sprintf("unexpected paths: %s, %s", repositoryFullFilepath, lsTreeEntry.FullFilepath))
			}

			submodulePath := filepath.ToSlash(submoduleFilepath)
			submoduleHandle, err := repoHandle.Submodule(submodulePath)
			if err != nil {
				return err
			}

			submoduleTree, err := getSubmoduleTree(submoduleHandle)
			if err != nil {
				return err
			}

			submoduleLsTreeEntrees, submoduleSubmoduleResults, err := lsTreeWalk(ctx, submoduleHandle, submoduleTree, lsTreeEntry.FullFilepath, lsTreeEntry.FullFilepath, opts)
			if err != nil {
				return err
			}

			if len(submoduleLsTreeEntrees) != 0 {
				submoduleName := submoduleHandle.Config().Name

				result := NewResult(submoduleHandle.Status().Expected.String(), lsTreeEntry.FullFilepath, submoduleLsTreeEntrees, submoduleSubmoduleResults)
				submoduleResult := NewSubmoduleResult(submoduleName, submodulePath, result)

				if !submoduleResult.IsEmpty() {
					submodulesResults = append(submodulesResults, submoduleResult)
				}
			}

			return nil
		},
		// skip tree func
		func() error {
			if debug() {
				logboek.Context(ctx).Debug().LogLn("Submodule entry was skipped: ", lsTreeEntry.FullFilepath)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return
}

func lsTreeDirOrSubmoduleEntryMatchBase(path string, opts LsTreeOptions, addTreeFunc, checkTreeFunc, skipTreeFunc func() error) error {
	pathMatcher := opts.formattedPathMatcher()
	switch {
	case pathMatcher.ShouldGoThrough(path):
		return checkTreeFunc()
	case pathMatcher.IsPathMatched(path):
		if opts.AllFiles {
			return checkTreeFunc()
		}

		return addTreeFunc()
	default:
		return skipTreeFunc()
	}
}

func lsTreeFileEntryMatch(ctx context.Context, lsTreeEntry *LsTreeEntry, opts LsTreeOptions) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if opts.formattedPathMatcher().IsPathMatched(lsTreeEntry.FullFilepath) {
		if debug() {
			logboek.Context(ctx).Debug().LogLn("File entry was added:        ", lsTreeEntry.FullFilepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	}

	return
}

func treeFindEntry(_ context.Context, tree repo_handle.TreeHandle, treeFullFilepath, treeEntryFilepath string) (*LsTreeEntry, error) {
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

func treeTree(tree repo_handle.TreeHandle, treeFullFilepath, treeDirEntryFullFilepath string) (repo_handle.TreeHandle, error) {
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

func getSubmoduleTree(repoHandle repo_handle.SubmoduleHandle) (repo_handle.TreeHandle, error) {
	submoduleName := repoHandle.Config().Name
	expectedCommit := repoHandle.Status().Expected
	tree, err := repoHandle.GetCommitTree(expectedCommit)
	if err != nil {
		return nil, fmt.Errorf("unable to get submodule %q commit %q tree: %w", submoduleName, expectedCommit, err)
	}

	return tree, nil
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_LS_TREE_PROCESS") == "1"
}
