package ls_tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/path_matcher"
)

func LsTree(repository *git.Repository, pathMatcher path_matcher.PathMatcher) (*Result, error) {
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
		repository:                       repository,
		tree:                             tree,
		lsTreeEntries:                    []*LsTreeEntry{},
		submodulesResults:                []*SubmoduleResult{},
		notInitializedSubmoduleFilepaths: []string{},
	}

	worktreeNotInitializedSubmodulePaths, err := notInitializedSubmoduleFilepaths(repository, "", pathMatcher)
	if err != nil {
		return nil, err
	}
	res.notInitializedSubmoduleFilepaths = worktreeNotInitializedSubmodulePaths

	baseFilepath := pathMatcher.BasePath()
	if baseFilepath != "" {
		basePathLsTreeEntry, err := treeFindEntry(tree, baseFilepath)
		if err != nil {
			if err == object.ErrDirectoryNotFound {
				return res, nil
			}

			return nil, err
		}

		lsTreeEntries, submodulesLsTreeEntries, err := lsTreeEntryMatch(repository, tree, "", basePathLsTreeEntry, pathMatcher)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = lsTreeEntries
		res.submodulesResults = submodulesLsTreeEntries

		return res, nil
	}

	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath("")
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Root tree was added")
		}

		rootTreeEntry := &LsTreeEntry{
			Filepath: "",
			TreeEntry: object.TreeEntry{
				Name: "",
				Mode: filemode.Dir,
				Hash: tree.Hash,
			},
		}

		res.lsTreeEntries = append(res.lsTreeEntries, rootTreeEntry)
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Root tree was checking")
		}

		lsTreeEntries, submodulesLsTreeEntries, err := lsTreeWalk(repository, tree, "", pathMatcher)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = lsTreeEntries
		res.submodulesResults = submodulesLsTreeEntries
	} else {
		if debugProcess() {
			logboek.Debug.LogLn("Root tree was skipped")
		}
	}

	return res, nil
}

func lsTreeWalk(repository *git.Repository, tree *object.Tree, treeFilepath string, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	for _, treeEntry := range tree.Entries {
		lsTreeEntry := &LsTreeEntry{
			Filepath:  filepath.Join(treeFilepath, treeEntry.Name),
			TreeEntry: treeEntry,
		}

		entryTreeEntries, entrySubmodulesTreeEntries, err := lsTreeEntryMatch(repository, tree, treeFilepath, lsTreeEntry, pathMatcher)
		if err != nil {
			return nil, nil, err
		}

		lsTreeEntries = append(lsTreeEntries, entryTreeEntries...)
		submodulesResults = append(submodulesResults, entrySubmodulesTreeEntries...)
	}

	return
}

func lsTreeEntryMatch(repository *git.Repository, tree *object.Tree, treeFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	switch lsTreeEntry.Mode {
	case filemode.Dir:
		return lsTreeDirEntryMatch(repository, tree, treeFilepath, lsTreeEntry, pathMatcher)
	case filemode.Submodule:
		return lsTreeSubmoduleEntryMatch(repository, tree, treeFilepath, lsTreeEntry, pathMatcher)
	default:
		return lsTreeFileEntryMatch(repository, tree, treeFilepath, lsTreeEntry, pathMatcher)
	}
}

func lsTreeDirEntryMatch(repository *git.Repository, tree *object.Tree, treeFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.Filepath)
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Dir entry was added:         ", lsTreeEntry.Filepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Dir entry was checking:      ", lsTreeEntry.Filepath)
		}
		entryTree, err := treeTree(tree, treeFilepath, lsTreeEntry.Filepath)
		if err != nil {
			return nil, nil, err
		}

		return lsTreeWalk(repository, entryTree, lsTreeEntry.Filepath, pathMatcher)
	}

	return
}

func lsTreeSubmoduleEntryMatch(repository *git.Repository, _ *object.Tree, _ string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.Filepath)
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Submodule entry was added:   ", lsTreeEntry.Filepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Submodule entry was checking:", lsTreeEntry.Filepath)
		}

		submoduleRepository, submoduleTree, err := submoduleRepositoryAndTree(repository, filepath.ToSlash(lsTreeEntry.Filepath))
		if err != nil {
			if err == git.ErrSubmoduleNotInitialized {
				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						lsTreeEntry.Filepath,
					)
				}

				return nil, nil, nil
			}
			return nil, nil, err
		}

		submoduleLsTreeEntrees, submoduleSubmoduleResults, err := lsTreeWalk(submoduleRepository, submoduleTree, lsTreeEntry.Filepath, pathMatcher)
		if err != nil {
			return nil, nil, err
		}

		if len(submoduleLsTreeEntrees) != 0 {
			submoduleResult := &SubmoduleResult{
				&Result{
					repository:                       submoduleRepository,
					tree:                             submoduleTree,
					treeFilepath:                     lsTreeEntry.Filepath,
					lsTreeEntries:                    submoduleLsTreeEntrees,
					submodulesResults:                submoduleSubmoduleResults,
					notInitializedSubmoduleFilepaths: []string{},
				},
			}

			if !submoduleResult.IsEmpty() {
				submodulesResults = append(submodulesResults, submoduleResult)
			}
		}
	}

	return
}

func lsTreeFileEntryMatch(_ *git.Repository, _ *object.Tree, _ string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if pathMatcher.MatchPath(lsTreeEntry.Filepath) {
		if debugProcess() {
			logboek.Debug.LogLn("File entry was added:        ", lsTreeEntry.Filepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	}

	return
}

func treeFindEntry(tree *object.Tree, treeEntryFilepath string) (*LsTreeEntry, error) {
	formattedTreeEntryPath := filepath.ToSlash(treeEntryFilepath)
	treeEntry, err := tree.FindEntry(formattedTreeEntryPath)
	if err != nil {
		return nil, err
	}

	return &LsTreeEntry{Filepath: treeEntryFilepath, TreeEntry: *treeEntry}, nil
}

func treeTree(tree *object.Tree, treeFilepath, treeDirEntryFilepath string) (*object.Tree, error) {
	entryFilepathRelativeToTree, err := filepath.Rel(treeFilepath, treeDirEntryFilepath)
	if err != nil || strings.HasPrefix(entryFilepathRelativeToTree, "..") {
		panicMsg := fmt.Sprintf(
			"runtime error: tree dir entry path (%s) must be subdirectory of tree path (%s)",
			treeFilepath,
			treeDirEntryFilepath,
		)

		if err != nil {
			panicMsg = fmt.Sprintf("%s: %s", panicMsg, err.Error())
		}

		panic(panicMsg)
	}

	formattedTreeDirEntryPath := filepath.ToSlash(entryFilepathRelativeToTree)
	entryTree, err := tree.Tree(formattedTreeDirEntryPath)
	if err != nil {
		return nil, err
	}

	return entryTree, nil
}

func notInitializedSubmoduleFilepaths(repository *git.Repository, relToBaseRepositoryFilepath string, pathMatcher path_matcher.PathMatcher) ([]string, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	var resultFilepaths []string
	for _, submodule := range submodules {
		submoduleFilepath := filepath.Join(relToBaseRepositoryFilepath, filepath.FromSlash(submodule.Config().Path))
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFilepath)
		if isMatched || shouldGoThrough {
			submoduleRepository, err := submodule.Repository()
			if err != nil {
				if err == git.ErrSubmoduleNotInitialized {
					resultFilepaths = append(resultFilepaths, submoduleFilepath)

					if debugProcess() {
						logboek.Debug.LogFWithCustomStyle(
							logboek.StyleByName(logboek.FailStyleName),
							"Submodule is not initialized: path %s will be added to checksum\n",
							relToBaseRepositoryFilepath,
						)
					}

					continue
				}

				return nil, err
			}

			submoduleFilepaths, err := notInitializedSubmoduleFilepaths(submoduleRepository, submoduleFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			resultFilepaths = append(resultFilepaths, submoduleFilepaths...)
		}
	}

	return resultFilepaths, nil
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

	submoduleStatus, err := submodule.Status()
	if err != nil {
		return nil, nil, err
	}

	if debugProcess() {
		if !submoduleStatus.IsClean() {
			logboek.Debug.LogFWithCustomStyle(
				logboek.StyleByName(logboek.FailStyleName),
				"Submodule is not clean (current commit %s), expected commit %s will be checked\n",
				submoduleStatus.Current,
				submoduleStatus.Expected,
			)
		}
	}

	commit, err := submoduleRepository.CommitObject(submoduleStatus.Expected)
	if err != nil {
		return nil, nil, err
	}

	submoduleTree, err := commit.Tree()
	if err != nil {
		return nil, nil, err
	}

	return submoduleRepository, submoduleTree, nil
}

func debugProcess() bool {
	return os.Getenv("WERF_DEBUG_LS_TREE_PROCESS") == "1"
}
