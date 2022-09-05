package git_repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"gopkg.in/ini.v1"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/git_repo/repo_handle"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/util/timestamps"
	"github.com/werf/werf/pkg/werf"
)

type Remote struct {
	*Base
	Url      string
	IsDryRun bool

	Endpoint *transport.Endpoint
}

func OpenRemoteRepo(name, url string) (*Remote, error) {
	repo := &Remote{Url: url}
	repo.Base = NewBase(name, repo.initRepoHandleBackedByWorkTree)
	return repo, repo.ValidateEndpoint()
}

func (repo *Remote) IsLocal() bool {
	return false
}

func (repo *Remote) GetWorkTreeDir() string {
	panic("not implemented")
}

func (repo *Remote) ValidateEndpoint() error {
	if ep, err := transport.NewEndpoint(repo.Url); err != nil {
		return fmt.Errorf("bad url %q: %w", repo.Url, err)
	} else {
		repo.Endpoint = ep
	}
	return nil
}

func (repo *Remote) CreateDetachedMergeCommit(ctx context.Context, fromCommit, toCommit string) (string, error) {
	return repo.createDetachedMergeCommit(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getWorkTreeCacheDir(repo.getRepoID()), fromCommit, toCommit)
}

func (repo *Remote) GetMergeCommitParents(_ context.Context, commit string) ([]string, error) {
	return repo.getMergeCommitParents(repo.GetClonePath(), commit)
}

func (repo *Remote) StatusPathList(ctx context.Context, pathMatcher path_matcher.PathMatcher) ([]string, error) {
	panic("not implemented")
}

func (repo *Remote) ValidateStatusResult(ctx context.Context, pathMatcher path_matcher.PathMatcher) error {
	panic("not implemented")
}

func (repo *Remote) getFilesystemRelativePathByEndpoint() string {
	host := repo.Endpoint.Host
	if repo.Endpoint.Port > 0 {
		host += fmt.Sprintf(":%d", repo.Endpoint.Port)
	}
	return filepath.Join(fmt.Sprintf("protocol-%s", repo.Endpoint.Protocol), host, repo.Endpoint.Path)
}

func (repo *Remote) GetClonePath() string {
	return filepath.Join(GetGitRepoCacheDir(), repo.getRepoID())
}

func (repo *Remote) RemoteOriginUrl(_ context.Context) (string, error) {
	return repo.remoteOriginUrl(repo.GetClonePath())
}

func (repo *Remote) IsEmpty(ctx context.Context) (bool, error) {
	return repo.isEmpty(ctx, repo.GetClonePath())
}

func (repo *Remote) IsShallowClone(ctx context.Context) (bool, error) {
	panic("not implemented")
}

func (repo *Remote) IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ctx, ancestorCommit, descendantCommit, repo.GetClonePath())
}

func (repo *Remote) CloneAndFetch(ctx context.Context) error {
	isCloned, err := repo.Clone(ctx)
	if err != nil {
		return err
	}
	if isCloned {
		return nil
	}

	return repo.FetchOrigin(ctx, FetchOptions{})
}

func (repo *Remote) isCloneExists() (bool, error) {
	_, err := os.Stat(repo.GetClonePath())
	if err == nil {
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, fmt.Errorf("cannot clone git repo: %w", err)
	}

	return false, nil
}

func (repo *Remote) updateLastAccessAt(ctx context.Context, repoPath string) error {
	path := filepath.Join(repoPath, "last_access_at")

	if _, lock, err := werf.AcquireHostLock(ctx, path, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("error locking path %q: %w", path, err)
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	return timestamps.WriteTimestampFile(path, time.Now())
}

func (repo *Remote) Clone(ctx context.Context) (bool, error) {
	if repo.IsDryRun {
		return false, nil
	}

	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return false, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	var err error

	exists, err := repo.isCloneExists()
	if err != nil {
		return false, err
	}
	if exists {
		if err := repo.updateLastAccessAt(ctx, repo.GetClonePath()); err != nil {
			return false, fmt.Errorf("error updating last access at timestamp: %w", err)
		}
		return false, nil
	}

	return true, repo.withRemoteRepoLock(ctx, func() error {
		exists, err := repo.isCloneExists()
		if err != nil {
			return err
		}
		if exists {
			if err := repo.updateLastAccessAt(ctx, repo.GetClonePath()); err != nil {
				return fmt.Errorf("error updating last access at timestamp: %w", err)
			}

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("Clone %s\n", repo.Url)

		if err := os.MkdirAll(filepath.Dir(repo.GetClonePath()), 0o755); err != nil {
			return fmt.Errorf("unable to create dir %s: %w", filepath.Dir(repo.GetClonePath()), err)
		}

		tmpPath := fmt.Sprintf("%s.tmp", repo.GetClonePath())
		// Remove previously created possibly existing dir
		if err := os.RemoveAll(tmpPath); err != nil {
			return fmt.Errorf("unable to prepare tmp path %s: failed to remove: %w", tmpPath, err)
		}
		// Ensure cleanup on failure
		defer os.RemoveAll(tmpPath)

		_, err = git.PlainClone(tmpPath, true, &git.CloneOptions{
			URL:               repo.Url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			return fmt.Errorf("unable to clone repo: %w", err)
		}

		if err := repo.updateLastAccessAt(ctx, tmpPath); err != nil {
			return fmt.Errorf("error updating last access at timestamp: %w", err)
		}

		if err := os.Rename(tmpPath, repo.GetClonePath()); err != nil {
			return fmt.Errorf("rename %s to %s failed: %w", tmpPath, repo.GetClonePath(), err)
		}

		return nil
	})
}

func (repo *Remote) SyncWithOrigin(ctx context.Context) error {
	panic("not implemented")
}

func (repo *Remote) Unshallow(ctx context.Context) error {
	panic("not implemented")
}

func (repo *Remote) FetchOrigin(ctx context.Context, opts FetchOptions) error {
	if repo.IsDryRun {
		return nil
	}

	cfgPath := filepath.Join(repo.GetClonePath(), "config")

	cfg, err := ini.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("cannot load repo %q config: %w", repo.String(), err)
	}

	remoteName := "origin"

	oldUrlKey := cfg.Section(fmt.Sprintf("remote \"%s\"", remoteName)).Key("url")
	if oldUrlKey != nil && oldUrlKey.Value() != repo.Url {
		oldUrlKey.SetValue(repo.Url)
		err := cfg.SaveTo(cfgPath)
		if err != nil {
			return fmt.Errorf("cannot update url of repo %q: %w", repo.String(), err)
		}
	}

	return repo.withRemoteRepoLock(ctx, func() error {
		rawRepo, err := git.PlainOpenWithOptions(repo.GetClonePath(), &git.PlainOpenOptions{EnableDotGitCommonDir: true})
		if err != nil {
			return fmt.Errorf("cannot open repo: %w", err)
		}

		logboek.Context(ctx).Default().LogFDetails("Fetch remote %s of %s\n", remoteName, repo.Url)

		err = rawRepo.Fetch(&git.FetchOptions{RemoteName: remoteName, Force: true, Tags: git.AllTags})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("cannot fetch remote %q of repo %q: %w", remoteName, repo.String(), err)
		}

		return nil
	})
}

func (repo *Remote) HeadCommitHash(ctx context.Context) (string, error) {
	return getHeadCommit(ctx, repo.GetClonePath())
}

func (repo *Remote) HeadCommitTime(ctx context.Context) (*time.Time, error) {
	time, err := baseHeadCommitTime(repo, ctx)
	return time, err
}

func (repo *Remote) findReference(rawRepo *git.Repository, reference string) (string, error) {
	refs, err := rawRepo.References()
	if err != nil {
		return "", err
	}

	var res string

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().String() == reference {
			res = ref.Hash().String()
			return storer.ErrStop
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return res, nil
}

func (repo *Remote) LatestBranchCommit(ctx context.Context, branch string) (string, error) {
	var err error

	rawRepo, err := git.PlainOpenWithOptions(repo.GetClonePath(), &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("cannot open repo: %w", err)
	}

	res, err := repo.findReference(rawRepo, fmt.Sprintf("refs/remotes/origin/%s", branch))
	if err != nil {
		return "", err
	}
	if res == "" {
		return "", fmt.Errorf("unknown branch %q of repo %q", branch, repo.String())
	}

	logboek.Context(ctx).Info().LogF("Using commit %q of repo %q branch %q\n", res, repo.String(), branch)

	return res, nil
}

func (repo *Remote) TagCommit(ctx context.Context, tag string) (string, error) {
	var err error

	rawRepo, err := git.PlainOpenWithOptions(repo.GetClonePath(), &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("cannot open repo: %w", err)
	}

	ref, err := rawRepo.Tag(tag)
	if err != nil {
		return "", fmt.Errorf("bad tag %q of repo %s: %w", tag, repo.String(), err)
	}

	var res string

	obj, err := rawRepo.TagObject(ref.Hash())
	switch err {
	case nil:
		// Tag object present
		res = obj.Target.String()
	case plumbing.ErrObjectNotFound:
		res = ref.Hash().String()
	default:
		return "", fmt.Errorf("bad tag %q of repo %s: %w", tag, repo.String(), err)
	}

	logboek.Context(ctx).Info().LogF("Using commit %q of repo %q tag %q\n", res, repo.String(), tag)

	return res, nil
}

func (repo *Remote) GetOrCreatePatch(ctx context.Context, opts PatchOptions) (Patch, error) {
	return repo.getOrCreatePatch(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getRepoID(), repo.getWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Remote) GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	return repo.getOrCreateArchive(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getRepoID(), repo.getWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Remote) GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (checksum string, err error) {
	err = repo.withRepoHandle(ctx, opts.Commit, func(repoHandle repo_handle.Handle) error {
		checksum, err = repo.getOrCreateChecksum(ctx, repoHandle, opts)
		return err
	})

	return
}

func (repo *Remote) IsCommitExists(ctx context.Context, commit string) (bool, error) {
	return repo.isCommitExists(ctx, repo.GetClonePath(), repo.GetClonePath(), commit)
}

func (repo *Remote) getRepoID() string {
	return util.Sha256Hash(repo.getFilesystemRelativePathByEndpoint())
}

func (repo *Remote) getWorkTreeCacheDir(repoID string) string {
	return filepath.Join(GetWorkTreeCacheDir(), "remote", repoID)
}

func (repo *Remote) withRemoteRepoLock(ctx context.Context, f func() error) error {
	lockName := fmt.Sprintf("remote_git_mapping.%s", repo.Name)
	return werf.WithHostLock(ctx, lockName, lockgate.AcquireOptions{Timeout: 600 * time.Second}, f)
}

func (repo *Remote) TagsList(_ context.Context) ([]string, error) {
	return repo.tagsList(repo.GetClonePath())
}

func (repo *Remote) RemoteBranchesList(_ context.Context) ([]string, error) {
	return repo.remoteBranchesList(repo.GetClonePath())
}

func (repo *Remote) initRepoHandleBackedByWorkTree(ctx context.Context, commit string) (repo_handle.Handle, error) {
	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	repository, err := git.PlainOpenWithOptions(repo.GetClonePath(), &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("cannot open git repository %q: %w", repo.GetClonePath(), err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %w", commit, err)
	}

	commitObj, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %q: %w", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return nil, err
	}

	var repoHandle repo_handle.Handle
	if err := true_git.WithWorkTree(ctx, repo.GetClonePath(), repo.getWorkTreeCacheDir(repo.getRepoID()), commit, true_git.WithWorkTreeOptions{HasSubmodules: hasSubmodules}, func(preparedWorkTreeDir string) error {
		repositoryWithPreparedWorktree, err := true_git.GitOpenWithCustomWorktreeDir(repo.GetClonePath(), preparedWorkTreeDir)
		if err != nil {
			return err
		}

		repoHandle, err = repo_handle.NewHandle(repositoryWithPreparedWorktree)
		return err
	}); err != nil {
		return nil, err
	}

	return repoHandle, nil
}
