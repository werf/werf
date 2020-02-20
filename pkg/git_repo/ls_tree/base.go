package ls_tree

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/flant/logboek"
)

func LsTree(repository *git.Repository, basePath string, pathFilter PathFilter) (*Result, error) {
	ref, err := repository.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repository.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	res := &Result{
		repository:        repository,
		tree:              tree,
		lsTreeEntries:     []*LsTreeEntry{},
		submodulesResults: []*SubmoduleResult{},
	}

	if basePath != "" {
		basePathLsTreeEntry, err := treeFindEntry(tree, basePath)
		if err != nil {
			if err == object.ErrDirectoryNotFound {
				return res, nil
			}

			return nil, err
		}

		lsTreeEntries, submodulesLsTreeEntries, err := lsTreeEntryMatch(repository, tree, "", basePathLsTreeEntry, pathFilter)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = lsTreeEntries
		res.submodulesResults = submodulesLsTreeEntries

		return res, nil
	}

	isTreeMatched, shouldWalkThrough := pathFilter.ProcessDirOrSubmodulePath("")
	if isTreeMatched {
		logboek.Debug.LogLn("Root tree was added")

		mainTreeEntry := &LsTreeEntry{
			Path: "",
			TreeEntry: object.TreeEntry{
				Name: "",
				Mode: filemode.Dir,
				Hash: tree.Hash,
			},
		}

		res.lsTreeEntries = append(res.lsTreeEntries, mainTreeEntry)
	} else if shouldWalkThrough {
		logboek.Debug.LogLn("Root tree was opened")
		lsTreeEntries, submodulesLsTreeEntries, err := lsTreeWalk(repository, tree, "", pathFilter)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = lsTreeEntries
		res.submodulesResults = submodulesLsTreeEntries
	} else {
		logboek.Debug.LogLn("Root tree was skipped")
	}

	return res, nil
}

func lsTreeWalk(repository *git.Repository, tree *object.Tree, treePath string, pathFilter PathFilter) ([]*LsTreeEntry, []*SubmoduleResult, error) {
	var lsTreeEntries []*LsTreeEntry
	var submodulesResults []*SubmoduleResult

	for _, treeEntry := range tree.Entries {
		lsTreeEntry := &LsTreeEntry{
			Path:      filepath.Join(treePath, treeEntry.Name),
			TreeEntry: treeEntry,
		}

		entryTreeEntries, entrySubmodulesTreeEntries, err := lsTreeEntryMatch(repository, tree, treePath, lsTreeEntry, pathFilter)
		if err != nil {
			return nil, nil, err
		}

		lsTreeEntries = append(lsTreeEntries, entryTreeEntries...)
		submodulesResults = append(submodulesResults, entrySubmodulesTreeEntries...)
	}

	return lsTreeEntries, submodulesResults, nil
}

func lsTreeEntryMatch(repository *git.Repository, tree *object.Tree, treePath string, lsTreeEntry *LsTreeEntry, pathFilter PathFilter) ([]*LsTreeEntry, []*SubmoduleResult, error) {
	switch lsTreeEntry.Mode {
	case filemode.Dir:
		return lsTreeDirEntryMatch(repository, tree, treePath, lsTreeEntry, pathFilter)
	case filemode.Submodule:
		return lsTreeSubmoduleEntryMatch(repository, tree, treePath, lsTreeEntry, pathFilter)
	default:
		return lsTreeFileEntryMatch(repository, tree, treePath, lsTreeEntry, pathFilter)
	}
}

func lsTreeDirEntryMatch(repository *git.Repository, tree *object.Tree, treePath string, lsTreeEntry *LsTreeEntry, pathFilter PathFilter) ([]*LsTreeEntry, []*SubmoduleResult, error) {
	var lsTreeEntries []*LsTreeEntry
	var submodulesResults []*SubmoduleResult

	isTreeMatched, shouldWalkThrough := pathFilter.ProcessDirOrSubmodulePath(lsTreeEntry.Path)
	if isTreeMatched {
		logboek.Debug.LogLn("Dir entry was added:        ", lsTreeEntry.Path)
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
		return lsTreeEntries, submodulesResults, nil
	} else if shouldWalkThrough {
		logboek.Debug.LogLn("Dir entry was opened:       ", lsTreeEntry.Path)

		entryTree, err := treeTree(tree, treePath, lsTreeEntry.Path)
		if err != nil {
			return nil, nil, err
		}

		return lsTreeWalk(repository, entryTree, lsTreeEntry.Path, pathFilter)
	} else {
		return lsTreeEntries, submodulesResults, nil
	}
}

func lsTreeSubmoduleEntryMatch(repository *git.Repository, _ *object.Tree, _ string, lsTreeEntry *LsTreeEntry, pathFilter PathFilter) ([]*LsTreeEntry, []*SubmoduleResult, error) {
	var lsTreeEntries []*LsTreeEntry
	var submoduleResults []*SubmoduleResult

	isTreeMatched, shouldWalkThrough := pathFilter.ProcessDirOrSubmodulePath(lsTreeEntry.Path)
	if isTreeMatched {
		logboek.Debug.LogLn("Submodule entry was added:  ", lsTreeEntry.Path)
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
		return lsTreeEntries, submoduleResults, nil
	} else if shouldWalkThrough {
		logboek.Debug.LogLn("Submodule entry was opened: ", lsTreeEntry.Path)

		submoduleRepository, submoduleTree, err := submoduleRepositoryAndTree(repository, lsTreeEntry.Path)
		if err != nil {
			return nil, nil, err
		}
		// TODO submodule is not initialized (dockerfile context)

		submoduleLsTreeEntrees, submoduleSubmoduleResults, err := lsTreeWalk(submoduleRepository, submoduleTree, lsTreeEntry.Path, pathFilter)

		if len(submoduleLsTreeEntrees) != 0 {
			submoduleResult := &SubmoduleResult{
				&Result{
					repository:        submoduleRepository,
					tree:              submoduleTree,
					treePath:          lsTreeEntry.Path,
					lsTreeEntries:     submoduleLsTreeEntrees,
					submodulesResults: submoduleSubmoduleResults,
				},
			}

			if !submoduleResult.empty() {
				submoduleResults = append(submoduleResults, submoduleResult)
			}
		}

		return lsTreeEntries, submoduleResults, nil
	} else {
		return lsTreeEntries, submoduleResults, nil
	}
}

func lsTreeFileEntryMatch(_ *git.Repository, _ *object.Tree, _ string, lsTreeEntry *LsTreeEntry, pathFilter PathFilter) ([]*LsTreeEntry, []*SubmoduleResult, error) {
	var lsTreeEntries []*LsTreeEntry
	var submodulesResults []*SubmoduleResult

	if pathFilter.MatchPath(lsTreeEntry.Path) {
		logboek.Debug.LogLn("File entry was added:       ", lsTreeEntry.Path)
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	}

	return lsTreeEntries, submodulesResults, nil
}

func treeFindEntry(tree *object.Tree, treeEntryPath string) (*LsTreeEntry, error) {
	formattedTreeEntryPath := filepath.ToSlash(treeEntryPath)
	treeEntry, err := tree.FindEntry(formattedTreeEntryPath)
	if err != nil {
		return nil, err
	}

	return &LsTreeEntry{Path: treeEntryPath, TreeEntry: *treeEntry}, nil
}

func treeTree(tree *object.Tree, treePath, treeDirEntryPath string) (*object.Tree, error) {
	entryPathRelativeToTree, err := filepath.Rel(treePath, treeDirEntryPath)
	if err != nil || strings.HasPrefix(entryPathRelativeToTree, "..") {
		panicMsg := fmt.Sprintf(
			"runtime error: tree dir entry path (%s) must be subdirectory of tree path (%s)",
			treePath,
			treeDirEntryPath,
		)

		if err != nil {
			panicMsg = fmt.Sprintf("%s: %s", panicMsg, err.Error())
		}

		panic(panicMsg)
	}

	formattedTreeDirEntryPath := filepath.ToSlash(entryPathRelativeToTree)
	entryTree, err := tree.Tree(formattedTreeDirEntryPath)
	if err != nil {
		return nil, err
	}

	return entryTree, nil
}

func submoduleRepositoryAndTree(repository *git.Repository, submoduleName string) (*git.Repository, *object.Tree, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, nil, err
	}

	submodule, err := worktree.Submodule(submoduleName)
	if err != nil {
		return nil, nil, err
	}

	submoduleRepository, err := submodule.Repository()
	if err != nil {
		return nil, nil, err
	}

	ref, err := submoduleRepository.Head()
	if err != nil {
		return nil, nil, err
	}

	commit, err := submoduleRepository.CommitObject(ref.Hash())
	if err != nil {
		return nil, nil, err
	}

	submoduleTree, err := commit.Tree()
	if err != nil {
		return nil, nil, err
	}

	return submoduleRepository, submoduleTree, nil
}
