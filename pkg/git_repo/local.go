package git_repo

import (
	"context"
	"fmt"
	"os"
	pathPkg "path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo/status"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

var ErrLocalRepositoryNotExists = git.ErrRepositoryNotExists

type Local struct {
	Base

	WorkTreeDir string
	GitDir      string

	headCommit string

	statusResult       *status.Result
	commitLsTreeResult map[string]*ls_tree.Result
	mutex              sync.Mutex
}

type OpenLocalRepoOptions struct {
	Dev bool
}

func OpenLocalRepo(name, workTreeDir string, opts OpenLocalRepoOptions) (l *Local, err error) {
	_, err = git.PlainOpenWithOptions(workTreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return l, ErrLocalRepositoryNotExists
		}

		return l, err
	}

	gitDir, err := true_git.ResolveRepoDir(filepath.Join(workTreeDir, git.GitDirName))
	if err != nil {
		return l, fmt.Errorf("unable to resolve git repo dir for %s: %s", workTreeDir, err)
	}

	l, err = newLocal(name, workTreeDir, gitDir)
	if err != nil {
		return l, err
	}

	if opts.Dev {
		devHeadCommit, err := true_git.SyncDevBranchWithStagedFiles(
			context.Background(),
			l.GitDir,
			l.getRepoWorkTreeCacheDir(l.getRepoID()),
			l.headCommit,
		)
		if err != nil {
			return l, err
		}

		l.headCommit = devHeadCommit
	}

	return l, nil
}

func newLocal(name, workTreeDir, gitDir string) (l *Local, err error) {
	headCommit, err := getHeadCommit(workTreeDir)
	if err != nil {
		return l, fmt.Errorf("unable to get git repo head commit: %s", err)
	}

	l = &Local{
		Base:               NewBase(name),
		WorkTreeDir:        workTreeDir,
		GitDir:             gitDir,
		headCommit:         headCommit,
		commitLsTreeResult: map[string]*ls_tree.Result{},
	}

	return l, nil
}

func (repo *Local) PlainOpen() (*git.Repository, error) {
	repository, err := git.PlainOpenWithOptions(repo.WorkTreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open git work tree %q: %s", repo.WorkTreeDir, err)
	}

	return repository, nil
}

func (repo *Local) SyncWithOrigin(ctx context.Context) error {
	isShallow, err := repo.IsShallowClone()
	if err != nil {
		return fmt.Errorf("check shallow clone failed: %s", err)
	}

	remoteOriginUrl, err := repo.RemoteOriginUrl(ctx)
	if err != nil {
		return fmt.Errorf("get remote origin failed: %s", err)
	}

	if remoteOriginUrl == "" {
		return fmt.Errorf("git remote origin was not detected")
	}

	return logboek.Context(ctx).Default().LogProcess("Syncing origin branches and tags").DoError(func() error {
		fetchOptions := true_git.FetchOptions{
			Prune:     true,
			PruneTags: true,
			Unshallow: isShallow,
			RefSpecs:  map[string]string{"origin": "+refs/heads/*:refs/remotes/origin/*"},
		}

		if err := true_git.Fetch(ctx, repo.WorkTreeDir, fetchOptions); err != nil {
			return fmt.Errorf("fetch failed: %s", err)
		}

		return nil
	})
}

func (repo *Local) FetchOrigin(ctx context.Context) error {
	isShallow, err := repo.IsShallowClone()
	if err != nil {
		return fmt.Errorf("check shallow clone failed: %s", err)
	}

	remoteOriginUrl, err := repo.RemoteOriginUrl(ctx)
	if err != nil {
		return fmt.Errorf("get remote origin failed: %s", err)
	}

	if remoteOriginUrl == "" {
		return fmt.Errorf("git remote origin was not detected")
	}

	return logboek.Context(ctx).Default().LogProcess("Fetching origin").DoError(func() error {
		fetchOptions := true_git.FetchOptions{
			Unshallow: isShallow,
			RefSpecs:  map[string]string{"origin": "+refs/heads/*:refs/remotes/origin/*"},
		}

		if err := true_git.Fetch(ctx, repo.WorkTreeDir, fetchOptions); err != nil {
			return fmt.Errorf("fetch failed: %s", err)
		}

		return nil
	})
}

func (repo *Local) IsShallowClone() (bool, error) {
	return true_git.IsShallowClone(repo.WorkTreeDir)
}

func (repo *Local) CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error) {
	return repo.createDetachedMergeCommit(ctx, repo.GitDir, repo.WorkTreeDir, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), fromCommit, toCommit)
}

func (repo *Local) GetMergeCommitParents(_ context.Context, commit string) ([]string, error) {
	return repo.getMergeCommitParents(repo.GitDir, commit)
}

type LsTreeOptions struct {
	Commit        string
	UseHeadCommit bool
	Strict        bool
}

func (repo *Local) LsTree(ctx context.Context, pathMatcher path_matcher.PathMatcher, opts LsTreeOptions) (*ls_tree.Result, error) {
	mainLsTreeResult, err := repo.getMainLsTreeResult(ctx, opts)
	if err != nil {
		return nil, err
	}

	var lsTreeResult *ls_tree.Result
	if err := repo.yieldRepositoryBackedByWorkTree(ctx, repo.headCommit, func(repository *git.Repository) (err error) {
		lsTreeResult, err = mainLsTreeResult.LsTree(ctx, repository, pathMatcher)
		return err
	}); err != nil {
		return nil, err
	}

	return lsTreeResult, nil
}

func (repo *Local) getMainLsTreeResult(ctx context.Context, opts LsTreeOptions) (*ls_tree.Result, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var commit string
	if opts.UseHeadCommit {
		commit = repo.headCommit
	} else if opts.Commit == "" {
		panic("no commit specified for LsTree procedure: specify Commit or HeadCommit")
	} else {
		commit = opts.Commit
	}

	_, ok := repo.commitLsTreeResult[commit]
	if ok {
		return repo.commitLsTreeResult[commit], nil
	}

	var lsTreeResult *ls_tree.Result
	if err := repo.yieldRepositoryBackedByWorkTree(ctx, commit, func(repository *git.Repository) error {
		r, err := ls_tree.LsTree(ctx, repository, commit, path_matcher.NewSimplePathMatcher("", []string{}, false), opts.Strict)
		if err != nil {
			return err
		}

		lsTreeResult = r

		return nil
	}); err != nil {
		return nil, err
	}

	repo.commitLsTreeResult[commit] = lsTreeResult
	return lsTreeResult, nil
}

func (repo *Local) Status(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*status.Result, error) {
	result, err := repo.getMainStatusResult(ctx)
	if err != nil {
		return nil, err
	}

	return result.Status(ctx, pathMatcher)
}

func (repo *Local) getMainStatusResult(ctx context.Context) (*status.Result, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if repo.statusResult == nil {
		if err := repo.InitAndSetMainStatusResult(ctx); err != nil {
			return nil, err
		}
	}

	return repo.statusResult, nil
}

func (repo *Local) InitAndSetMainStatusResult(ctx context.Context) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("InitAndSetMainStatusResult").
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			err = repo.initAndSetMainStatusResult(ctx)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (repo *Local) initAndSetMainStatusResult(ctx context.Context) error {
	repository, err := repo.PlainOpen()
	if err != nil {
		return err
	}

	result, err := status.Status(ctx, repository, path_matcher.NewSimplePathMatcher("", []string{}, true))
	if err != nil {
		return err
	}

	repo.statusResult = result
	return nil
}

func (repo *Local) IsEmpty(ctx context.Context) (bool, error) {
	return repo.isEmpty(ctx, repo.WorkTreeDir)
}

func (repo *Local) IsAncestor(_ context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ancestorCommit, descendantCommit, repo.GitDir)
}

func (repo *Local) RemoteOriginUrl(ctx context.Context) (string, error) {
	return repo.remoteOriginUrl(repo.WorkTreeDir)
}

func (repo *Local) HeadCommit(_ context.Context) (string, error) {
	return repo.headCommit, nil
}

func (repo *Local) GetOrCreatePatch(ctx context.Context, opts PatchOptions) (Patch, error) {
	return repo.getOrCreatePatch(ctx, repo.WorkTreeDir, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	return repo.getOrCreateArchive(ctx, repo.WorkTreeDir, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (checksum Checksum, checksumErr error) {
	return repo.getOrCreateChecksum(ctx, repo.yieldRepositoryBackedByWorkTree, opts)
}

func (repo *Local) IsCommitExists(ctx context.Context, commit string) (bool, error) {
	return repo.isCommitExists(ctx, repo.WorkTreeDir, repo.GitDir, commit)
}

func (repo *Local) TagsList(ctx context.Context) ([]string, error) {
	return repo.tagsList(repo.WorkTreeDir)
}

func (repo *Local) RemoteBranchesList(ctx context.Context) ([]string, error) {
	return repo.remoteBranchesList(repo.WorkTreeDir)
}

func (repo *Local) getRepoID() string {
	absPath, err := filepath.Abs(repo.WorkTreeDir)
	if err != nil {
		panic(err) // stupid interface of filepath.Abs
	}

	fullPath := filepath.Clean(absPath)
	return util.Sha256Hash(fullPath)
}

func (repo *Local) getRepoWorkTreeCacheDir(repoID string) string {
	return filepath.Join(GetWorkTreeCacheDir(), "local", repoID)
}

func (repo *Local) ValidateSubmodules(ctx context.Context, matcher path_matcher.PathMatcher) error {
	mainStatusResult, err := repo.getMainStatusResult(ctx)
	if err != nil {
		return err
	}

	statusResult, err := mainStatusResult.Status(ctx, matcher)
	if err != nil {
		return err
	}

	repository, err := repo.PlainOpen()
	if err != nil {
		return err
	}

	return statusResult.ValidateSubmodules(repository, repo.headCommit)
}

// IsFileModifiedLocally checks if the file has worktree or staged changes
func (repo *Local) IsFileModifiedLocally(ctx context.Context, path string, options status.FilterOptions) (bool, error) {
	statusResult, err := repo.getMainStatusResult(ctx)
	if err != nil {
		return false, err
	}

	isModified := statusResult.IsFileModified(path, options)

	return isModified, nil
}

func (repo *Local) GetModifiedLocallyFilePathList(ctx context.Context, matcher path_matcher.PathMatcher, options status.FilterOptions) ([]string, error) {
	mainStatusResult, err := repo.getMainStatusResult(ctx)
	if err != nil {
		return nil, err
	}

	statusResult, err := mainStatusResult.Status(ctx, matcher)
	if err != nil {
		return nil, err
	}

	return statusResult.FilePathList(options), nil
}

// ListCommitFilesWithGlob returns the list of files by the glob, follows symlinks.
// The result paths are relative to the passed directory, the method does reverse resolving for symlinks.
func (repo *Local) ListCommitFilesWithGlob(ctx context.Context, commit string, dir string, glob string) (files []string, err error) {
	var prefixWithoutPatterns string
	prefixWithoutPatterns, glob = util.GlobPrefixWithoutPatterns(glob)
	dirOrFileWithGlobPrefix := filepath.Join(dir, prefixWithoutPatterns)

	pathMatcher := path_matcher.NewSimplePathMatcher(dirOrFileWithGlobPrefix, []string{glob}, true)
	if debugGiterminismManager() {
		logboek.Context(ctx).Debug().LogLn("pathMatcher:", pathMatcher.String())
	}

	var result []string
	fileFunc := func(notResolvedPath string) error {
		if pathMatcher.MatchPath(notResolvedPath) {
			result = append(result, notResolvedPath)
		}

		return nil
	}

	isRegularFile, err := repo.IsCommitFileExist(ctx, commit, dirOrFileWithGlobPrefix)
	if err != nil {
		return nil, err
	}

	if isRegularFile {
		if err := fileFunc(dirOrFileWithGlobPrefix); err != nil {
			return nil, err
		}

		return result, nil
	}

	if err := repo.WalkCommitFiles(ctx, commit, dirOrFileWithGlobPrefix, pathMatcher, fileFunc); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Local) WalkCommitFiles(ctx context.Context, commit string, dir string, pathMatcher path_matcher.PathMatcher, fileFunc func(notResolvedPath string) error) error {
	isDirMatched, shouldGoThroughDir := pathMatcher.ProcessDirOrSubmodulePath(dir)
	possiblyDirMatched := isDirMatched || shouldGoThroughDir
	if !possiblyDirMatched {
		return nil
	}

	exist, err := repo.IsCommitDirectoryExist(ctx, commit, dir)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	resolvedDir, err := repo.ResolveCommitFilePath(ctx, commit, dir)
	if err != nil {
		return fmt.Errorf("unable to resolve commit file %q: %s", dir, err)
	}

	result, err := repo.LsTree(ctx, path_matcher.NewSimplePathMatcher(resolvedDir, []string{}, true), LsTreeOptions{
		Commit: commit,
		Strict: true,
	})
	if err != nil {
		return err
	}

	return result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		notResolvedPath := strings.Replace(filepath.ToSlash(lsTreeEntry.FullFilepath), resolvedDir, dir, 1)

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- %q %q\n", notResolvedPath, lsTreeEntry.Mode.String())
		}

		if lsTreeEntry.Mode.IsMalformed() {
			panic(fmt.Sprintf("unexpected condition: %+v", lsTreeEntry))
		}

		isMatched, shouldGoThrough := pathMatcher.ProcessDirOrSubmodulePath(notResolvedPath)
		possiblyMatched := isMatched || shouldGoThrough
		if !possiblyMatched {
			return nil
		}

		if lsTreeEntry.Mode == filemode.Symlink {
			isDir, err := repo.IsCommitDirectoryExist(ctx, commit, notResolvedPath)
			if err != nil {
				return err
			}

			if isDir {
				err := repo.WalkCommitFiles(ctx, commit, notResolvedPath, pathMatcher, fileFunc)
				if err != nil {
					return err
				}
			} else {
				if err := fileFunc(notResolvedPath); err != nil {
					return err
				}
			}
		} else {
			if err := fileFunc(notResolvedPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// ReadCommitFile resolves symlinks and returns commit tree entry content.
func (repo *Local) ReadCommitFile(ctx context.Context, commit, path string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadCommitFile %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = repo.readCommitFile(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	return
}

func (repo *Local) readCommitFile(ctx context.Context, commit, path string) ([]byte, error) {
	resolvedPath, err := repo.ResolveCommitFilePath(ctx, commit, path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve commit file %q: %s", path, err)
	}

	return repo.ReadCommitTreeEntryContent(ctx, commit, resolvedPath)
}

// IsCommitFileExist resolves symlinks and returns true if the resolved commit tree entry is Regular, Deprecated, or Executable.
func (repo *Local) IsCommitFileExist(ctx context.Context, commit, path string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitFileExist %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isCommitFileExist(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return

}

func (repo *Local) isCommitFileExist(ctx context.Context, commit, path string) (bool, error) {
	return repo.checkCommitFileMode(ctx, commit, path, []filemode.FileMode{filemode.Regular, filemode.Deprecated, filemode.Executable})
}

// IsCommitDirectoryExist resolves symlinks and returns true if the resolved commit tree entry is Dir or Submodule.
func (repo *Local) IsCommitDirectoryExist(ctx context.Context, commit, path string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitDirectoryExist %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isCommitDirectoryExist(ctx, commit, path)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (repo *Local) isCommitDirectoryExist(ctx context.Context, commit, path string) (bool, error) {
	return repo.checkCommitFileMode(ctx, commit, path, []filemode.FileMode{filemode.Dir, filemode.Submodule})
}

func (repo *Local) checkCommitFileMode(ctx context.Context, commit string, path string, expectedFileModeList []filemode.FileMode) (bool, error) {
	resolvedPath, err := repo.ResolveCommitFilePath(ctx, commit, path)
	if err != nil {
		if IsTreeEntryNotFoundInRepoErr(err) {
			return false, nil
		}

		return false, fmt.Errorf("unable to resolve commit file %q: %s", path, err)
	}

	lsTreeEntry, err := repo.getCommitTreeEntry(ctx, commit, resolvedPath)
	if err != nil {
		return false, fmt.Errorf("unable to get commit tree entry %q: %s", resolvedPath, err)
	}

	for _, mode := range expectedFileModeList {
		if mode == lsTreeEntry.Mode {
			return true, nil
		}
	}

	return false, nil
}

// ResolveAndCheckCommitFilePath does ResolveCommitFilePath with an additional check for each resolved link target.
func (repo *Local) ResolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(resolvedPath string) error) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveAndCheckCommitFilePath %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			checkWithDebugFunc := func(resolvedPath string) error {
				return logboek.Context(ctx).Debug().
					LogBlock("-- check %s", resolvedPath).
					Options(func(options types.LogBlockOptionsInterface) {
						if !debugGiterminismManager() {
							options.Mute()
						}
					}).
					DoError(func() error {
						err := checkSymlinkTargetFunc(resolvedPath)

						if debugGiterminismManager() {
							logboek.Context(ctx).Debug().LogF("err: %q\n", err)
						}

						return err
					})
			}

			resolvedPath, err = repo.resolveAndCheckCommitFilePath(ctx, commit, path, checkWithDebugFunc)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %v\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

func (repo *Local) resolveAndCheckCommitFilePath(ctx context.Context, commit, path string, checkSymlinkTargetFunc func(relPath string) error) (resolvedPath string, err error) {
	return repo.resolveCommitFilePath(ctx, commit, path, 0, checkSymlinkTargetFunc)
}

// ResolveCommitFilePath follows symbolic links and returns the resolved path if there is a corresponding tree entry in the repo.
func (repo *Local) ResolveCommitFilePath(ctx context.Context, commit, path string) (resolvedPath string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ResolveCommitFilePath %q %q", commit, path).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			resolvedPath, err = repo.resolveCommitFilePath(ctx, commit, path, 0, nil)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("resolvedPath: %v\nerr: %q\n", resolvedPath, err)
			}
		})

	return
}

type treeEntryNotFoundInRepoErr struct {
	error
}

func IsTreeEntryNotFoundInRepoErr(err error) bool {
	switch err.(type) {
	case treeEntryNotFoundInRepoErr:
		return true
	default:
		return false
	}
}

func (repo *Local) resolveCommitFilePath(ctx context.Context, commit, path string, depth int, checkSymlinkTargetFunc func(resolvedPath string) error) (string, error) {
	if depth > 1000 {
		return "", fmt.Errorf("too many levels of symbolic links")
	}
	depth++

	// returns path if the corresponding tree entry is Regular, Deprecated, Executable, Dir, or Submodule.
	{
		lsTreeEntry, err := repo.getCommitTreeEntry(ctx, commit, path)
		if err != nil {
			return "", fmt.Errorf("unable to get commit tree entry %q: %s", path, err)
		}

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- [*] %s %s %s\n", path, lsTreeEntry.Mode.String(), err)
		}

		if lsTreeEntry.Mode != filemode.Symlink && !lsTreeEntry.Mode.IsMalformed() {
			return path, nil
		}
	}

	pathParts := util.SplitFilepath(path)
	pathPartsLen := len(pathParts)

	var resolvedPath string
	for ind := 0; ind < pathPartsLen; ind++ {
		pathToResolve := pathPkg.Join(resolvedPath, pathParts[ind])

		lsTreeEntry, err := repo.getCommitTreeEntry(ctx, commit, pathToResolve)
		if err != nil {
			return "", fmt.Errorf("unable to get commit tree entry %q: %s", pathToResolve, err)
		}

		if debugGiterminismManager() {
			logboek.Context(ctx).Debug().LogF("-- [%d] %s %s %s\n", ind, pathToResolve, lsTreeEntry.Mode.String(), err)
		}

		mode := lsTreeEntry.Mode
		switch {
		case mode.IsMalformed():
			return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", pathToResolve)}
		case mode == filemode.Symlink:
			data, err := repo.ReadCommitTreeEntryContent(ctx, commit, pathToResolve)
			if err != nil {
				return "", fmt.Errorf("unable to get commit tree entry content %q: %s", pathToResolve, err)
			}

			link := string(data)
			if pathPkg.IsAbs(link) {
				return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", link)}
			}

			resolvedLink := pathPkg.Join(pathPkg.Dir(pathToResolve), link)
			if resolvedLink == ".." || strings.HasPrefix(resolvedLink, "../") {
				return "", treeEntryNotFoundInRepoErr{fmt.Errorf("commit tree entry %q not found in the repository", link)}
			}

			if checkSymlinkTargetFunc != nil {
				if err := checkSymlinkTargetFunc(resolvedLink); err != nil {
					return "", err
				}
			}

			resolvedTarget, err := repo.resolveCommitFilePath(ctx, commit, resolvedLink, depth, checkSymlinkTargetFunc)
			if err != nil {
				return "", err
			}

			resolvedPath = resolvedTarget
		default:
			resolvedPath = pathToResolve
		}
	}

	return resolvedPath, nil
}

func (repo *Local) ReadCommitTreeEntryContent(ctx context.Context, commit, relPath string) ([]byte, error) {
	lsTreeResult, err := repo.LsTree(ctx, path_matcher.NewSimplePathMatcher(relPath, []string{}, false), LsTreeOptions{
		Commit: commit,
		Strict: true,
	})
	if err != nil {
		return nil, err
	}

	var content []byte
	if err := repo.yieldRepositoryBackedByWorkTree(ctx, commit, func(repository *git.Repository) error {
		content, err = lsTreeResult.LsTreeEntryContent(repository, relPath)
		return err
	}); err != nil {
		return nil, err
	}

	return content, nil
}

func (repo *Local) IsCommitTreeEntryDirectory(ctx context.Context, commit, relPath string) (isDirectory bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitTreeEntryDirectory %q %q", commit, relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			isDirectory, err = repo.isCommitTreeEntryDirectory(ctx, commit, relPath)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("isDirectory: %v\nerr: %q\n", isDirectory, err)
			}
		})

	return
}

func (repo *Local) isCommitTreeEntryDirectory(ctx context.Context, commit, relPath string) (bool, error) {
	entry, err := repo.getCommitTreeEntry(ctx, commit, relPath)
	if err != nil {
		return false, err
	}

	return entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule, nil
}

func (repo *Local) IsCommitTreeEntryExist(ctx context.Context, commit, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsCommitTreeEntryExist %q %q", commit, relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = repo.isTreeEntryExist(ctx, commit, relPath)

			if debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (repo *Local) isTreeEntryExist(ctx context.Context, commit, relPath string) (bool, error) {
	entry, err := repo.getCommitTreeEntry(ctx, commit, relPath)
	if err != nil {
		return false, err
	}

	return !entry.Mode.IsMalformed(), nil
}

func (repo *Local) getCommitTreeEntry(ctx context.Context, commit, path string) (*ls_tree.LsTreeEntry, error) {
	lsTreeResult, err := repo.LsTree(ctx, path_matcher.NewSimplePathMatcher(path, []string{}, false), LsTreeOptions{
		Commit: commit,
		Strict: true,
	})
	if err != nil {
		return nil, err
	}

	entry := lsTreeResult.LsTreeEntry(path)

	return entry, nil
}

func (repo *Local) yieldRepositoryBackedByWorkTree(ctx context.Context, commit string, doFunc func(repository *git.Repository) error) error {
	repository, err := repo.PlainOpen()
	if err != nil {
		return err
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("bad commit %q: %s", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return err
	}

	if hasSubmodules {
		return true_git.WithWorkTree(ctx, repo.GitDir, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), commit, true_git.WithWorkTreeOptions{HasSubmodules: hasSubmodules}, func(preparedWorkTreeDir string) error {
			repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(repo.GitDir, preparedWorkTreeDir)
			if err != nil {
				return err
			}

			return doFunc(repositoryWithPreparedWorktree)
		})
	} else {
		return doFunc(repository)
	}
}

func debugGiterminismManager() bool {
	return os.Getenv("WERF_DEBUG_GITERMINISM_MANAGER") == "1"
}
