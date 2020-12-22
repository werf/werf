package git_repo

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/werf/logboek"

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

	headCommit string
}

func OpenLocalRepo(name, path string, dev bool) (*Local, error) {
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

	localRepo, err := newLocal(name, path, gitDir)
	if err != nil {
		return nil, err
	}

	if dev {
		devHeadCommit, err := true_git.SyncDevBranchWithStagedFiles(
			context.Background(),
			localRepo.GitDir,
			localRepo.getRepoWorkTreeCacheDir(localRepo.getRepoID()),
			localRepo.headCommit,
		)
		if err != nil {
			return nil, err
		}

		localRepo.headCommit = devHeadCommit
	}

	return localRepo, nil
}

func newLocal(name, path, gitDir string) (*Local, error) {
	headCommit, err := getHeadCommit(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get git repo head commit: %s", err)
	}

	local := &Local{
		Base:       Base{Name: name},
		Path:       path,
		GitDir:     gitDir,
		headCommit: headCommit,
	}

	return local, nil
}

func (repo *Local) PlainOpen() (*git.Repository, error) {
	return git.PlainOpen(repo.Path)
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

		if err := true_git.Fetch(ctx, repo.Path, fetchOptions); err != nil {
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

		if err := true_git.Fetch(ctx, repo.Path, fetchOptions); err != nil {
			return fmt.Errorf("fetch failed: %s", err)
		}

		return nil
	})
}

func (repo *Local) IsShallowClone() (bool, error) {
	return true_git.IsShallowClone(repo.Path)
}

func (repo *Local) CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error) {
	return repo.createDetachedMergeCommit(ctx, repo.GitDir, repo.Path, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), fromCommit, toCommit)
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
	repository, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.Path, err)
	}

	var commit string
	if opts.UseHeadCommit {
		commit = repo.headCommit
	} else if opts.Commit == "" {
		panic(fmt.Sprintf("no commit specified for LsTree procedure: specify Commit or HeadCommit"))
	} else {
		commit = opts.Commit
	}

	return ls_tree.LsTree(ctx, repository, commit, pathMatcher, opts.Strict)
}

func (repo *Local) Status(ctx context.Context, pathMatcher path_matcher.PathMatcher) (*status.Result, error) {
	repository, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.Path, err)
	}

	return status.Status(ctx, repository, repo.Path, pathMatcher)
}

func (repo *Local) IsEmpty(ctx context.Context) (bool, error) {
	return repo.isEmpty(ctx, repo.Path)
}

func (repo *Local) IsAncestor(_ context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ancestorCommit, descendantCommit, repo.GitDir)
}

func (repo *Local) RemoteOriginUrl(ctx context.Context) (string, error) {
	return repo.remoteOriginUrl(repo.Path)
}

func (repo *Local) HeadCommit(_ context.Context) (string, error) {
	return repo.headCommit, nil
}

func (repo *Local) CreatePatch(ctx context.Context, opts PatchOptions) (Patch, error) {
	return repo.createPatch(ctx, repo.Path, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) CreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(ctx, repo.Path, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) Checksum(ctx context.Context, opts ChecksumOptions) (checksum Checksum, err error) {
	logboek.Context(ctx).Debug().LogProcess("Calculating checksum").Do(func() {
		checksum, err = repo.checksumWithLsTree(ctx, repo.Path, repo.GitDir, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
	})

	return checksum, err
}

func (repo *Local) CheckAndReadCommitSymlink(ctx context.Context, path string, commit string) (bool, []byte, error) {
	return repo.checkAndReadSymlink(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.Path, repo.GitDir, commit, path)
}

func (repo *Local) IsCommitExists(ctx context.Context, commit string) (bool, error) {
	return repo.isCommitExists(ctx, repo.Path, repo.GitDir, commit)
}

func (repo *Local) TagsList(ctx context.Context) ([]string, error) {
	return repo.tagsList(repo.Path)
}

func (repo *Local) RemoteBranchesList(ctx context.Context) ([]string, error) {
	return repo.remoteBranchesList(repo.Path)
}

func (repo *Local) getRepoID() string {
	absPath, err := filepath.Abs(repo.Path)
	if err != nil {
		panic(err) // stupid interface of filepath.Abs
	}

	fullPath := filepath.Clean(absPath)
	return util.Sha256Hash(fullPath)
}

func (repo *Local) getRepoWorkTreeCacheDir(repoID string) string {
	return filepath.Join(GetWorkTreeCacheDir(), "local", repoID)
}

func (repo *Local) IsCommitFileExists(ctx context.Context, commit, path string) (bool, error) {
	return repo.isCommitFileExists(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.Path, repo.GitDir, commit, path)
}

func (repo *Local) IsCommitDirectoryExists(ctx context.Context, dir string, commit string) (bool, error) {
	if paths, err := repo.GetCommitFilePathList(ctx, commit); err != nil {
		return false, fmt.Errorf("unable to get file path list from the local git repo commit %s: %s", commit, err)
	} else {
		cleanDirPath := filepath.ToSlash(filepath.Clean(dir))
		for _, path := range paths {
			isSubpath := util.IsSubpathOfBasePath(cleanDirPath, path)
			if isSubpath {
				return true, nil
			}
		}

		return false, nil
	}
}

func (repo *Local) GetCommitFilePathList(ctx context.Context, commit string) ([]string, error) {
	result, err := repo.LsTree(ctx, path_matcher.NewGitMappingPathMatcher("", nil, nil, true), LsTreeOptions{
		Commit: commit,
		Strict: true,
	})
	if err != nil {
		return nil, err
	}

	var res []string
	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		res = append(res, lsTreeEntry.FullFilepath)
		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *Local) ReadCommitFile(ctx context.Context, commit, path string) ([]byte, error) {
	return repo.readFile(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.Path, repo.GitDir, commit, path)
}
