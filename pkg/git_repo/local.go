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

var ErrLocalRepositoryNotExists = git.ErrRepositoryNotExists

type Local struct {
	Base

	WorkTreeDir string
	GitDir      string

	headCommit string
}

type OpenLocalRepoOptions struct {
	Dev bool
}

func OpenLocalRepo(name, workTreeDir string, opts OpenLocalRepoOptions) (l Local, err error) {
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

func newLocal(name, workTreeDir, gitDir string) (l Local, err error) {
	headCommit, err := getHeadCommit(workTreeDir)
	if err != nil {
		return l, fmt.Errorf("unable to get git repo head commit: %s", err)
	}

	l = Local{
		Base:        Base{Name: name},
		WorkTreeDir: workTreeDir,
		GitDir:      gitDir,
		headCommit:  headCommit,
	}

	return l, nil
}

func (repo *Local) PlainOpen() (*git.Repository, error) {
	return git.PlainOpen(repo.WorkTreeDir)
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
	repository, err := git.PlainOpenWithOptions(repo.WorkTreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.WorkTreeDir, err)
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
	repository, err := git.PlainOpenWithOptions(repo.WorkTreeDir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %s: %s", repo.WorkTreeDir, err)
	}

	return status.Status(ctx, repository, repo.WorkTreeDir, pathMatcher)
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

func (repo *Local) CreatePatch(ctx context.Context, opts PatchOptions) (Patch, error) {
	return repo.createPatch(ctx, repo.WorkTreeDir, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) CreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	return repo.createArchive(ctx, repo.WorkTreeDir, repo.GitDir, repo.getRepoID(), repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Local) Checksum(ctx context.Context, opts ChecksumOptions) (checksum Checksum, err error) {
	logboek.Context(ctx).Debug().LogProcess("Calculating checksum").Do(func() {
		checksum, err = repo.checksumWithLsTree(ctx, repo.WorkTreeDir, repo.GitDir, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), opts)
	})

	return checksum, err
}

func (repo *Local) CheckAndReadCommitSymlink(ctx context.Context, path string, commit string) (bool, []byte, error) {
	return repo.checkAndReadSymlink(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.WorkTreeDir, repo.GitDir, commit, path)
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

func (repo *Local) IsCommitFileExists(ctx context.Context, commit, path string) (bool, error) {
	return repo.isCommitFileExists(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.WorkTreeDir, repo.GitDir, commit, path)
}

func (repo *Local) IsCommitDirectoryExists(ctx context.Context, commit string, dir string) (bool, error) {
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
	return repo.readFile(ctx, repo.getRepoWorkTreeCacheDir(repo.getRepoID()), repo.WorkTreeDir, repo.GitDir, commit, path)
}
