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
		repository:                           repository,
		repositoryFullFilepath:               "",
		tree:                                 tree,
		lsTreeEntries:                        []*LsTreeEntry{},
		submodulesResults:                    []*SubmoduleResult{},
		notInitializedSubmoduleFullFilepaths: []string{},
	}

	worktreeNotInitializedSubmodulePaths, err := notInitializedSubmoduleFullFilepaths(repository, "", pathMatcher)
	if err != nil {
		return nil, err
	}
	res.notInitializedSubmoduleFullFilepaths = worktreeNotInitializedSubmodulePaths

	baseFilepath := pathMatcher.BaseFilepath()
	if baseFilepath != "" {
		lsTreeEntries, submodulesResults, err := processSpecificEntryFilepath(repository, tree, "", "", pathMatcher.BaseFilepath(), pathMatcher)
		if err != nil {
			return nil, err
		}

		res.lsTreeEntries = append(res.lsTreeEntries, lsTreeEntries...)
		res.submodulesResults = append(res.submodulesResults, submodulesResults...)

		return res, nil
	}

	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath("")
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Root tree was added")
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
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Root tree was checking")
		}

		lsTreeEntries, submodulesLsTreeEntries, err := lsTreeWalk(repository, tree, "", "", pathMatcher)
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

func processSpecificEntryFilepath(repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath, treeEntryFilepath string, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
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

		submoduleRepository, submoduleTree, err := submoduleRepositoryAndTree(repository, submodule.Config().Name)
		if err != nil {
			if err == git.ErrSubmoduleNotInitialized {
				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						submoduleFullFilepath,
					)
				}

				return lsTreeEntries, submodulesResults, nil
			}

			return nil, nil, err
		}

		sLsTreeEntries, sSubmodulesResults, err := processSpecificEntryFilepath(submoduleRepository, submoduleTree, submoduleFullFilepath, submoduleFullFilepath, relTreeEntryFilepath, pathMatcher)
		if err != nil {
			return nil, nil, err
		}

		submoduleResult := &SubmoduleResult{Result: &Result{
			repository:                           submoduleRepository,
			repositoryFullFilepath:               submoduleFullFilepath,
			tree:                                 submoduleTree,
			lsTreeEntries:                        sLsTreeEntries,
			submodulesResults:                    sSubmodulesResults,
			notInitializedSubmoduleFullFilepaths: []string{},
		}}

		if !submoduleResult.IsEmpty() {
			submodulesResults = append(submodulesResults, submoduleResult)
		}

		return lsTreeEntries, submodulesResults, nil
	}

	lsTreeEntry, err := treeFindEntry(tree, treeFullFilepath, treeEntryFilepath)
	if err != nil {
		if err == object.ErrDirectoryNotFound || err == object.ErrFileNotFound {
			return lsTreeEntries, submodulesResults, nil
		}

		return nil, nil, err
	}

	lsTreeEntries, submodulesLsTreeEntries, err := lsTreeEntryMatch(repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher)
	if err != nil {
		return nil, nil, err
	}

	return lsTreeEntries, submodulesLsTreeEntries, nil
}

func lsTreeWalk(repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	for _, treeEntry := range tree.Entries {
		lsTreeEntry := &LsTreeEntry{
			FullFilepath: filepath.Join(treeFullFilepath, treeEntry.Name),
			TreeEntry:    treeEntry,
		}

		entryTreeEntries, entrySubmodulesTreeEntries, err := lsTreeEntryMatch(repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher)
		if err != nil {
			return nil, nil, err
		}

		lsTreeEntries = append(lsTreeEntries, entryTreeEntries...)
		submodulesResults = append(submodulesResults, entrySubmodulesTreeEntries...)
	}

	return
}

func lsTreeEntryMatch(repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	switch lsTreeEntry.Mode {
	case filemode.Dir:
		return lsTreeDirEntryMatch(repository, tree, repositoryFullFilepath, treeFullFilepath, lsTreeEntry, pathMatcher)
	case filemode.Submodule:
		return lsTreeSubmoduleEntryMatch(repository, repositoryFullFilepath, lsTreeEntry, pathMatcher)
	default:
		return lsTreeFileEntryMatch(lsTreeEntry, pathMatcher)
	}
}

func lsTreeDirEntryMatch(repository *git.Repository, tree *object.Tree, repositoryFullFilepath, treeFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.FullFilepath)
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Dir entry was added:         ", lsTreeEntry.FullFilepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Dir entry was checking:      ", lsTreeEntry.FullFilepath)
		}
		entryTree, err := treeTree(tree, treeFullFilepath, lsTreeEntry.FullFilepath)
		if err != nil {
			return nil, nil, err
		}

		return lsTreeWalk(repository, entryTree, repositoryFullFilepath, lsTreeEntry.FullFilepath, pathMatcher)
	}

	return
}

func lsTreeSubmoduleEntryMatch(repository *git.Repository, repositoryFullFilepath string, lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	isTreeMatched, shouldWalkThrough := pathMatcher.ProcessDirOrSubmodulePath(lsTreeEntry.FullFilepath)
	if isTreeMatched {
		if debugProcess() {
			logboek.Debug.LogLn("Submodule entry was added:   ", lsTreeEntry.FullFilepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	} else if shouldWalkThrough {
		if debugProcess() {
			logboek.Debug.LogLn("Submodule entry was checking:", lsTreeEntry.FullFilepath)
		}

		submoduleFilepath, err := filepath.Rel(repositoryFullFilepath, lsTreeEntry.FullFilepath)
		if err != nil || submoduleFilepath == "." || submoduleFilepath == ".." || strings.HasPrefix(submoduleFilepath, ".."+string(os.PathSeparator)) {
			panic(fmt.Sprintf("unexpected paths: %s, %s", repositoryFullFilepath, lsTreeEntry.FullFilepath))
		}

		submoduleName := filepath.ToSlash(submoduleFilepath)
		submoduleRepository, submoduleTree, err := submoduleRepositoryAndTree(repository, submoduleName)
		if err != nil {
			if err == git.ErrSubmoduleNotInitialized {
				if debugProcess() {
					logboek.Debug.LogFWithCustomStyle(
						logboek.StyleByName(logboek.FailStyleName),
						"Submodule is not initialized: path %s will be added to checksum\n",
						lsTreeEntry.FullFilepath,
					)
				}

				return nil, nil, nil
			}
			return nil, nil, err
		}

		submoduleLsTreeEntrees, submoduleSubmoduleResults, err := lsTreeWalk(submoduleRepository, submoduleTree, lsTreeEntry.FullFilepath, lsTreeEntry.FullFilepath, pathMatcher)
		if err != nil {
			return nil, nil, err
		}

		if len(submoduleLsTreeEntrees) != 0 {
			submoduleResult := &SubmoduleResult{
				&Result{
					repository:                           submoduleRepository,
					repositoryFullFilepath:               lsTreeEntry.FullFilepath,
					tree:                                 submoduleTree,
					lsTreeEntries:                        submoduleLsTreeEntrees,
					submodulesResults:                    submoduleSubmoduleResults,
					notInitializedSubmoduleFullFilepaths: []string{},
				},
			}

			if !submoduleResult.IsEmpty() {
				submodulesResults = append(submodulesResults, submoduleResult)
			}
		}
	}

	return
}

func lsTreeFileEntryMatch(lsTreeEntry *LsTreeEntry, pathMatcher path_matcher.PathMatcher) (lsTreeEntries []*LsTreeEntry, submodulesResults []*SubmoduleResult, err error) {
	if pathMatcher.MatchPath(lsTreeEntry.FullFilepath) {
		if debugProcess() {
			logboek.Debug.LogLn("File entry was added:        ", lsTreeEntry.FullFilepath)
		}
		lsTreeEntries = append(lsTreeEntries, lsTreeEntry)
	}

	return
}

func treeFindEntry(tree *object.Tree, treeFullFilepath, treeEntryFilepath string) (*LsTreeEntry, error) {
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

func notInitializedSubmoduleFullFilepaths(repository *git.Repository, repositoryFullFilepath string, pathMatcher path_matcher.PathMatcher) ([]string, error) {
	worktree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := worktree.Submodules()
	if err != nil {
		return nil, err
	}

	var resultFullFilepaths []string
	for _, submodule := range submodules {
		submoduleEntryFilepath := filepath.FromSlash(submodule.Config().Path)
		submoduleFullFilepath := filepath.Join(repositoryFullFilepath, submoduleEntryFilepath)
		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(submoduleFullFilepath)
		if isMatched || shouldGoThrough {
			submoduleRepository, err := submodule.Repository()
			if err != nil {
				if err == git.ErrSubmoduleNotInitialized {
					resultFullFilepaths = append(resultFullFilepaths, submoduleFullFilepath)

					if debugProcess() {
						logboek.Debug.LogFWithCustomStyle(
							logboek.StyleByName(logboek.FailStyleName),
							"Submodule is not initialized: path %s will be added to checksum\n",
							submoduleFullFilepath,
						)
					}

					continue
				}

				return nil, err
			}

			submoduleFullFilepaths, err := notInitializedSubmoduleFullFilepaths(submoduleRepository, submoduleFullFilepath, pathMatcher)
			if err != nil {
				return nil, err
			}

			resultFullFilepaths = append(resultFullFilepaths, submoduleFullFilepaths...)
		}
	}

	return resultFullFilepaths, nil
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
