package git_repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
)

type Remote struct {
	*Base
	Url      string
	IsDryRun bool

	Endpoint *transport.Endpoint

	BasicAuth *BasicAuth
}

func OpenRemoteRepo(name, url string, auth *BasicAuthCredentials) (*Remote, error) {
	repo := &Remote{Url: url}
	repo.Base = NewBase(name, repo.initRepoHandleBackedByWorkTree)
	if auth != nil {
		basicAuth, err := BasicAuthCredentialsHelper(auth)
		if err != nil {
			return nil, fmt.Errorf("unable to get basic auth for repository %s: %w", name, err)
		}
		repo.BasicAuth = basicAuth
	}
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
		rawRepo, err := repo.PlainOpen()
		if err != nil {
			return fmt.Errorf("open cloned repo: %w", err)
		}
		return repo.syncLocalBranches(ctx, rawRepo)
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

	if _, lock, err := werf.HostLocker().AcquireLock(ctx, path, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("error locking path %q: %w", path, err)
	} else {
		defer werf.HostLocker().ReleaseLock(lock)
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
		defer werf.HostLocker().ReleaseLock(lock)
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

		cloneOpts := &git.CloneOptions{
			URL:               repo.Url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}

		if repo.BasicAuth != nil {
			cloneOpts.Auth = newBasicAuth(repo.BasicAuth.Username, repo.BasicAuth.Password).AuthMethod
		}

		_, err = git.PlainCloneContext(ctx, tmpPath, true, cloneOpts)
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

type Auth struct {
	AuthMethod transport.AuthMethod
}

type BasicAuth struct {
	Username string
	Password string
}

func newBasicAuth(username, password string) *Auth {
	return &Auth{
		AuthMethod: &http.BasicAuth{
			Username: username,
			Password: password,
		},
	}
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
		rawRepo, err := repo.PlainOpen()
		if err != nil {
			return fmt.Errorf("cannot open repo: %w", err)
		}

		logboek.Context(ctx).Default().LogFDetails("Fetch remote %s of %s\n", remoteName, repo.Url)

		fetchOpts := &git.FetchOptions{
			RemoteName: remoteName,
			Force:      true,
			Tags:       git.AllTags,
		}

		if repo.BasicAuth != nil {
			fetchOpts.Auth = newBasicAuth(repo.BasicAuth.Username, repo.BasicAuth.Password).AuthMethod
		}

		err = rawRepo.FetchContext(ctx, fetchOpts)
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("cannot fetch remote %q of repo %q: %w", remoteName, repo.String(), err)
		}

		if err := repo.syncLocalBranches(ctx, rawRepo); err != nil {
			return fmt.Errorf("cannot update local branches of repo %q: %w", repo.String(), err)
		}

		return nil
	})
}

func (repo *Remote) syncLocalBranches(ctx context.Context, rawRepo *git.Repository) error {
	if err := logboek.Context(ctx).Debug().LogProcess("Updating local branches").DoError(func() error {
		refs, err := rawRepo.References()
		if err != nil {
			return fmt.Errorf("cannot get references of repo %q: %w", repo.String(), err)
		}

		return refs.ForEach(func(ref *plumbing.Reference) error {
			name := ref.Name().String()
			if strings.HasPrefix(name, "refs/remotes/origin/") {
				branch := strings.TrimPrefix(name, "refs/remotes/origin/")
				localRefName := plumbing.ReferenceName("refs/heads/" + branch)

				if err := rawRepo.Storer.SetReference(plumbing.NewHashReference(localRefName, ref.Hash())); err != nil {
					return err
				}

				logboek.Context(ctx).Debug().LogLnDetails(branch, "->", ref.Hash())
			}
			return nil
		})
	}); err != nil {
		return fmt.Errorf("sync local branches: %w", err)
	}
	return nil
}

func (repo *Remote) PlainOpen() (*git.Repository, error) {
	return gitRepoPlainOpen(repo.GetClonePath())
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

	rawRepo, err := repo.PlainOpen()
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

	rawRepo, err := repo.PlainOpen()
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
	repository, err := repo.PlainOpen()
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %w", repo.GetClonePath(), err)
	}

	toHash, err := newHash(opts.ToCommit)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit hash %q: %w", opts.ToCommit, err)
	}

	toCommit, err := repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit %q: %w", opts.ToCommit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(toCommit)
	if err != nil {
		return nil, err
	}

	repoID := repo.getRepoID()
	workTreeCacheDir := repo.getWorkTreeCacheDir(repoID)
	if !hasSubmodules {
		return repo.getOrCreatePatch(ctx, repo.GetClonePath(), repo.GetClonePath(), repoID, workTreeCacheDir, opts)
	}

	patchID := true_git.PatchOptions(opts).ID()
	checksumMutex := util.MapLoadOrCreateMutex(&repo.Cache.patchesMutex, patchID)
	checksumMutex.Lock()
	defer checksumMutex.Unlock()

	if val, ok := repo.Cache.Patches.Load(patchID); ok {
		return val.(Patch), nil
	}

	lock, err := CommonGitDataManager.LockGC(ctx, true)
	if err != nil {
		return nil, err
	}
	defer werf.HostLocker().ReleaseLock(lock)

	if patch, err := CommonGitDataManager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patch != nil {
		repo.Cache.Patches.Store(patchID, patch)
		return patch, nil
	}

	tmpFile, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", tmpFile, err)
	}

	var retryCount int
	gitDir := repo.GetClonePath()

TryCreatePatch:
	desc, err := true_git.Patch(ctx, fileHandler, gitDir, workTreeCacheDir, true, true_git.PatchOptions(opts))
	if true_git.IsCommitsNotPresentError(err) && retryCount == 0 {
		logboek.Context(ctx).Default().LogF("Detected not present commits when creating patch: %s\n", err)
		logboek.Context(ctx).Default().LogF("Will switch worktree to original commit %q and retry\n", opts.FromCommit)
		if err := fileHandler.Truncate(0); err != nil {
			return nil, fmt.Errorf("unable to truncate file %s: %w", tmpFile, err)
		}
		if _, err := fileHandler.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("unable to reset file %s: %w", tmpFile, err)
		}
		if err := true_git.WithWorkTree(ctx, gitDir, workTreeCacheDir, opts.FromCommit, true_git.WithWorkTreeOptions{HasSubmodules: true}, func(workTreeDir string) error {
			return nil
		}); err != nil {
			return nil, fmt.Errorf("unable to switch worktree to commit %q: %w", opts.FromCommit, err)
		}
		retryCount++
		goto TryCreatePatch
	}

	if err != nil {
		return nil, fmt.Errorf("error creating patch between %q and %q commits: %w", opts.FromCommit, opts.ToCommit, err)
	}

	if err := fileHandler.Close(); err != nil {
		return nil, fmt.Errorf("error creating patch file %s: %w", tmpFile, err)
	}

	patch, err := CommonGitDataManager.CreatePatchFile(ctx, repoID, opts, tmpFile, desc)
	if err != nil {
		return nil, err
	}

	repo.Cache.Patches.Store(patchID, patch)

	return patch, nil
}

func (repo *Remote) GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	repository, err := repo.PlainOpen()
	if err != nil {
		return nil, fmt.Errorf("cannot open repo %q: %w", repo.GetClonePath(), err)
	}

	commitHash, err := newHash(opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("bad commit hash %q: %w", opts.Commit, err)
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit %q: %w", opts.Commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commit)
	if err != nil {
		return nil, err
	}

	repoID := repo.getRepoID()
	workTreeCacheDir := repo.getWorkTreeCacheDir(repoID)
	if !hasSubmodules {
		return repo.getOrCreateArchive(ctx, repo.GetClonePath(), repo.GetClonePath(), repoID, workTreeCacheDir, opts)
	}

	repo.Cache.archivesMutex.Lock()
	defer repo.Cache.archivesMutex.Unlock()

	archiveID := true_git.ArchiveOptions(opts).ID()
	if archive, hasKey := repo.Cache.Archives[archiveID]; hasKey {
		return archive, nil
	}

	lock, err := CommonGitDataManager.LockGC(ctx, true)
	if err != nil {
		return nil, err
	}
	defer werf.HostLocker().ReleaseLock(lock)

	if archive, err := CommonGitDataManager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archive != nil {
		repo.Cache.Archives[archiveID] = archive
		return archive, nil
	}

	tmpFile, err := CommonGitDataManager.NewTmpFile()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file: %w", err)
	}

	gitDir := repo.GetClonePath()
	err = true_git.ArchiveWithSubmodules(ctx, fileHandler, gitDir, workTreeCacheDir, true_git.ArchiveOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("error creating archive for commit %q: %w", opts.Commit, err)
	}

	if err := fileHandler.Close(); err != nil {
		return nil, fmt.Errorf("unable to close file %s: %w", tmpFile, err)
	}

	archive, err := CommonGitDataManager.CreateArchiveFile(ctx, repoID, opts, tmpFile)
	if err != nil {
		return nil, err
	}

	repo.Cache.Archives[archiveID] = archive

	return archive, nil
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
	return werf.HostLocker().WithLock(ctx, lockName, lockgate.AcquireOptions{Timeout: 600 * time.Second}, f)
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
		defer werf.HostLocker().ReleaseLock(lock)
	}

	repository, err := repo.PlainOpen()
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
