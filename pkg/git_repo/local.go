package git_repo

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo/check_ignore"
	"github.com/werf/werf/pkg/git_repo/status"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

type Local struct {
	Base
	Path   string
	GitDir string
}

func OpenLocalRepo(name string, path string) (*Local, error) {
	_, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return nil, nil
		}

		return nil, err
	}

	gitDir, err := true_git.GetRealRepoDir(filepath.Join(path, ".git"))
	if err != nil {
		return nil, fmt.Errorf("unable to get real git repo dir for %s: %s", path, err)
	}

	localRepo := &Local{Base: Base{Name: name}, Path: path, GitDir: gitDir}

	return localRepo, nil
}

func (repo *Local) PlainOpen() (*git.Repository, error) {
	return git.PlainOpen(repo.Path)
}

func (repo *Local) SyncWithOrigin() error {
	isShallow, err := repo.IsShallowClone()
	if err != nil {
		return fmt.Errorf("check shallow clone failed: %s", err)
	}

	remoteOriginUrl, err := repo.RemoteOriginUrl()
	if err != nil {
		return fmt.Errorf("get remote origin failed: %s", err)
	}

	if remoteOriginUrl == "" {
		return fmt.Errorf("git remote origin was not detected")
	}

	return logboek.Default.LogProcess("Syncing origin branches and tags", logboek.LevelLogProcessOptions{}, func() error {
		fetchOptions := true_git.FetchOptions{
			Prune:     true,
			PruneTags: true,
			Unshallow: isShallow,
			RefSpecs:  map[string]string{"origin": "+refs/heads/*:refs/remotes/origin/*"},
		}

		if err := true_git.Fetch(repo.Path, fetchOptions); err != nil {
			return fmt.Errorf("fetch failed: %s", err)
		}

		return nil
	})
}

func (repo *Local) FetchOrigin() error {
	isShallow, err := repo.IsShallowClone()
	if err != nil {
		return fmt.Errorf("check shallow clone failed: %s", err)
	}

	remoteOriginUrl, err := repo.RemoteOriginUrl()
	if err != nil {
		return fmt.Errorf("get remote origin failed: %s", err)
	}

	if remoteOriginUrl == "" {
		return fmt.Errorf("git remote origin was not detected")
	}

	return logboek.Default.LogProcess("Fetching origin", logboek.LevelLogProcessOptions{}, func() error {
		fetchOptions := true_git.FetchOptions{
			Unshallow: isShallow,
			RefSpecs:  map[string]string{"origin": "+refs/heads/*:refs/remotes/origin/*"},
		}

		if err := true_git.Fetch(repo.Path, fetchOptions); err != nil {
			return fmt.Errorf("fetch failed: %s", err)
		}

		return nil
	})
}

func (repo *Local) IsShallowClone() (bool, error) {
	return true_git.IsShallowClone(repo.Path)
}

func (repo *Local) CreateDetachedMergeCommit(fromCommit, toCommit string) (string, error) {
	return repo.createDetachedMergeCommit(repo.GitDir, repo.Path, repo.getRepoWorkTreeCacheDir(), fromCommit, toCommit)
}

func (repo *Local) GetMergeCommitParents(commit string) ([]string, error) {
	return repo.getMergeCommitParents(repo.GitDir, commit)
}

type LsTreeOptions struct {
	Commit        string
	UseHeadCommit bool
	Strict        bool
}

func (repo *Local) LsTree(pathMatcher path_matcher.PathMatcher, opts LsTreeOptions) (*ls_tree.Result, error) {
	repository, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.Path, err)
	}

	var commit string
	if opts.UseHeadCommit {
		if headCommit, err := repo.HeadCommit(); err != nil {
			return nil, fmt.Errorf("unable to get repo head commit: %s", err)
		} else {
			commit = headCommit
		}
	} else if opts.Commit == "" {
		panic(fmt.Sprintf("no commit specified for LsTree procedure: specify Commit or HeadCommit"))
	} else {
		commit = opts.Commit
	}

	return ls_tree.LsTree(repository, commit, pathMatcher, opts.Strict)
}

func (repo *Local) Status(pathMatcher path_matcher.PathMatcher) (*status.Result, error) {
	repository, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.Path, err)
	}

	return status.Status(repository, repo.Path, pathMatcher)
}

func (repo *Local) CheckIgnore(paths []string) (*check_ignore.Result, error) {
	repository, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.Path, err)
	}

	return check_ignore.CheckIgnore(repository, repo.Path, paths)
}

func (repo *Local) IsEmpty() (bool, error) {
	return repo.isEmpty(repo.Path)
}

func (repo *Local) IsAncestor(ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ancestorCommit, descendantCommit, repo.GitDir)
}

func (repo *Local) RemoteOriginUrl() (string, error) {
	return repo.remoteOriginUrl(repo.Path)
}

func (repo *Local) HeadCommit() (string, error) {
	return repo.getHeadCommit(repo.Path)
}

func (repo *Local) IsHeadReferenceExist() (bool, error) {
	_, err := repo.getHeadCommit(repo.Path)
	if err == errHeadNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (repo *Local) CreatePatch(opts PatchOptions) (Patch, error) {
	return repo.createPatch(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
}

func (repo *Local) CreateArchive(opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
}

func (repo *Local) Checksum(opts ChecksumOptions) (checksum Checksum, err error) {
	_ = logboek.Debug.LogProcess(
		"Calculating checksum",
		logboek.LevelLogProcessOptions{},
		func() error {
			checksum, err = repo.checksumWithLsTree(repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(), opts)
			return nil
		},
	)

	return
}

func (repo *Local) IsCommitExists(commit string) (bool, error) {
	return repo.isCommitExists(repo.Path, repo.GitDir, commit)
}

func (repo *Local) TagsList() ([]string, error) {
	return repo.tagsList(repo.Path)
}

func (repo *Local) RemoteBranchesList() ([]string, error) {
	return repo.remoteBranchesList(repo.Path)
}

func (repo *Local) getRepoWorkTreeCacheDir() string {
	absPath, err := filepath.Abs(repo.Path)
	if err != nil {
		panic(err) // stupid interface of filepath.Abs
	}

	fullPath := filepath.Clean(absPath)
	repoId := util.Sha256Hash(fullPath)

	return filepath.Join(GetWorkTreeCacheDir(), "local", repoId)
}
