package git_repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/status"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var ErrLocalRepositoryNotExists = git.ErrRepositoryNotExists

type Local struct {
	*Base

	WorkTreeDir string
	GitDir      string

	headCommit string

	statusResult *status.Result
	mutex        sync.Mutex
}

type OpenLocalRepoOptions struct {
	WithServiceHeadCommit    bool
	ServiceHeadCommitOptions ServiceHeadCommit
}

type ServiceHeadCommit struct {
	WithStagedChangesOnly bool // all tracked files if false
}

func OpenLocalRepo(ctx context.Context, name, workTreeDir string, opts OpenLocalRepoOptions) (l *Local, err error) {
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

	if opts.WithServiceHeadCommit {
		if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
			return nil, err
		} else {
			defer werf.ReleaseHostLock(lock)
		}

		devHeadCommit, err := true_git.SyncSourceWorktreeWithServiceWorktreeBranch(
			context.Background(),
			l.GitDir,
			l.WorkTreeDir,
			l.getRepoWorkTreeCacheDir(l.getRepoID()),
			l.headCommit,
			true_git.SyncSourceWorktreeWithServiceWorktreeBranchOptions{
				OnlyStagedChanges: opts.ServiceHeadCommitOptions.WithStagedChangesOnly,
			},
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
		WorkTreeDir: workTreeDir,
		GitDir:      gitDir,
		headCommit:  headCommit,
	}
	l.Base = NewBase(name, l.initRepoHandleBackedByWorkTree)

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

func (repo *Local) status(ctx context.Context) (*status.Result, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if repo.statusResult == nil {
		result, err := status.Status(ctx, repo.WorkTreeDir)
		if err != nil {
			return nil, err
		}

		repo.statusResult = &result
	}

	return repo.statusResult, nil
}

func (repo *Local) IsEmpty(ctx context.Context) (bool, error) {
	return repo.isEmpty(ctx, repo.WorkTreeDir)
}

func (repo *Local) IsAncestor(_ context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ancestorCommit, descendantCommit, repo.GitDir)
}

func (repo *Local) RemoteOriginUrl(_ context.Context) (string, error) {
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

func (repo *Local) GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (checksum string, err error) {
	err = repo.withRepoHandle(ctx, opts.Commit, func(repoHandle repo_handle.Handle) error {
		checksum, err = repo.getOrCreateChecksum(ctx, repoHandle, opts)
		return err
	})

	return
}

func (repo *Local) IsCommitExists(ctx context.Context, commit string) (bool, error) {
	return repo.isCommitExists(ctx, repo.WorkTreeDir, repo.GitDir, commit)
}

func (repo *Local) TagsList(_ context.Context) ([]string, error) {
	return repo.tagsList(repo.WorkTreeDir)
}

func (repo *Local) RemoteBranchesList(_ context.Context) ([]string, error) {
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

type UntrackedFilesFoundError StatusFilesFoundError
type UncommittedFilesFoundError StatusFilesFoundError
type StatusFilesFoundError struct {
	PathList []string
	error
}

type SubmoduleAddedAndNotCommittedError SubmoduleErrorBase
type SubmoduleDeletedError SubmoduleErrorBase
type SubmoduleHasUntrackedChangesError SubmoduleErrorBase
type SubmoduleHasUncommittedChangesError SubmoduleErrorBase
type SubmoduleCommitChangedError SubmoduleErrorBase
type SubmoduleErrorBase struct {
	SubmodulePath string
	error
}

type ValidateStatusResultOptions StatusPathListOptions

func (repo *Local) ValidateStatusResult(ctx context.Context, pathMatcher path_matcher.PathMatcher, opts ValidateStatusResultOptions) error {
	statusResult, err := repo.status(ctx)
	if err != nil {
		return err
	}

	var untrackedPathList []string
	for _, path := range statusResult.UntrackedPathList {
		if pathMatcher.IsPathMatched(path) {
			untrackedPathList = append(untrackedPathList, path)
		}
	}

	if len(untrackedPathList) != 0 {
		return UntrackedFilesFoundError{
			PathList: untrackedPathList,
			error:    fmt.Errorf("untracked files found"),
		}
	}

	if opts.OnlyUntrackedChanges {
		return nil
	}

	var scope status.Scope
	if opts.OnlyWorktreeChanges {
		scope = statusResult.Worktree
	} else {
		scope = statusResult.IndexWithWorktree()
	}

	var uncommittedPathList []string
	for _, path := range scope.PathList() {
		if pathMatcher.IsPathMatched(path) {
			uncommittedPathList = append(uncommittedPathList, path)
		}
	}

	if len(uncommittedPathList) != 0 {
		return UncommittedFilesFoundError{
			PathList: uncommittedPathList,
			error:    fmt.Errorf("uncommitted files found"),
		}
	}

	return repo.validateStatusResultSubmodules(ctx, pathMatcher, scope)
}

func (repo *Local) validateStatusResultSubmodules(_ context.Context, pathMatcher path_matcher.PathMatcher, scope status.Scope) error {
	// No changes related to submodules.
	if len(scope.Submodules()) == 0 {
		return nil
	}

	for _, submodule := range scope.Submodules() {
		if !pathMatcher.IsDirOrSubmodulePathMatched(submodule.Path) {
			continue
		}

		if submodule.IsAdded {
			return SubmoduleAddedAndNotCommittedError{
				SubmodulePath: submodule.Path,
				error:         fmt.Errorf("submodule is added but not committed"),
			}
		} else if submodule.IsDeleted {
			return SubmoduleDeletedError{
				SubmodulePath: submodule.Path,
				error:         fmt.Errorf("submodule is deleted"),
			}
		} else if submodule.IsModified {
			if submodule.HasUntrackedChanges {
				return SubmoduleHasUntrackedChangesError{
					SubmodulePath: submodule.Path,
					error:         fmt.Errorf("submodule has untracked changes"),
				}
			}

			if submodule.HasTrackedChanges {
				return SubmoduleHasUncommittedChangesError{
					SubmodulePath: submodule.Path,
					error:         fmt.Errorf("submodule has uncommitted changes"),
				}
			}

			if submodule.IsCommitChanged {
				return SubmoduleCommitChangedError{
					SubmodulePath: submodule.Path,
					error:         fmt.Errorf("submodule commit is changed"),
				}
			}
		}
	}

	return nil
}

type StatusPathListOptions struct {
	OnlyWorktreeChanges  bool
	OnlyUntrackedChanges bool
}

func (repo *Local) StatusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher, opts StatusPathListOptions) (list []string, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("StatusPathList %q %v", pathMatcher.String(), opts).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debugGiterminismManager() {
				options.Mute()
			}
		}).
		Do(func() {
			list, err = repo.statusPathList(ctx, pathMatcher, opts)

			if !debugGiterminismManager() {
				logboek.Context(ctx).Debug().LogLn("list:", list)
				logboek.Context(ctx).Debug().LogLn("err:", err)
			}
		})

	return
}

func (repo *Local) statusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher, opts StatusPathListOptions) ([]string, error) {
	statusResult, err := repo.status(ctx)
	if err != nil {
		return nil, err
	}

	var result []string
	handlePathListFunc := func(pathList []string) {
		for _, path := range pathList {
			if pathMatcher.IsPathMatched(path) {
				result = util.AddNewStringsToStringArray(result, path)
			}
		}
	}

	handlePathListFunc(statusResult.UntrackedPathList)
	if opts.OnlyUntrackedChanges {
		return result, nil
	}

	var scope status.Scope
	if opts.OnlyWorktreeChanges {
		scope = statusResult.Worktree
	} else {
		scope = statusResult.IndexWithWorktree()
	}
	handlePathListFunc(scope.PathList())

	for _, submodule := range scope.Submodules() {
		if pathMatcher.IsDirOrSubmodulePathMatched(submodule.Path) {
			result = util.AddNewStringsToStringArray(result, submodule.Path)
		}
	}

	return result, nil
}

func (repo *Local) StatusIndexChecksum(ctx context.Context) (string, error) {
	statusResult, err := repo.status(ctx)
	if err != nil {
		return "", err
	}

	return statusResult.Index.Checksum(), nil
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

func (repo *Local) initRepoHandleBackedByWorkTree(ctx context.Context, commit string) (repo_handle.Handle, error) {
	repository, err := repo.PlainOpen()
	if err != nil {
		return nil, err
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %s", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %q: %s", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return nil, err
	}

	if hasSubmodules {
		if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
			return nil, err
		} else {
			defer werf.ReleaseHostLock(lock)
		}

		var repoHandle repo_handle.Handle
		if err := true_git.WithWorkTree(ctx, repo.GitDir, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), commit, true_git.WithWorkTreeOptions{HasSubmodules: hasSubmodules}, func(preparedWorkTreeDir string) error {
			repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(repo.GitDir, preparedWorkTreeDir)
			if err != nil {
				return err
			}

			repoHandle, err = repo_handle.NewHandle(repositoryWithPreparedWorktree)
			return err
		}); err != nil {
			return nil, err
		}

		return repoHandle, nil
	} else {
		return repo_handle.NewHandle(repository)
	}
}

func debugGiterminismManager() bool {
	return os.Getenv("WERF_DEBUG_GITERMINISM_MANAGER") == "1"
}
