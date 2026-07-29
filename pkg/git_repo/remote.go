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
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
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

	Branch string
	Tag    string
	Commit string

	kind mirrorKind
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
	return repo.clonePathForKind(repo.mirrorKind())
}

func (repo *Remote) clonePathForKind(kind mirrorKind) string {
	if kind == mirrorKindFull {
		return filepath.Join(GetGitRepoCacheDir(), repo.getRepoID())
	}
	return filepath.Join(GetGitMirrorsCacheDir(), repo.getRepoID(), string(kind))
}

func (repo *Remote) mirrorKind() mirrorKind {
	if repo.kind == "" {
		return mirrorKindFull
	}
	return repo.kind
}

func (repo *Remote) requiresFullMarkerPath() string {
	return filepath.Join(GetGitMirrorsCacheDir(), repo.getRepoID(), "requires_full")
}

func (repo *Remote) resolveMirrorKind() (mirrorKind, error) {
	if repo.Tag == "" && repo.Commit == "" {
		return mirrorKindFull, nil
	}

	markerExists, err := util.FileExists(repo.requiresFullMarkerPath())
	if err != nil {
		return "", fmt.Errorf("unable to check requires_full marker %q: %w", repo.requiresFullMarkerPath(), err)
	}
	if markerExists {
		return mirrorKindFull, nil
	}

	return mirrorKindShallow, nil
}

func (repo *Remote) RemoteOriginUrl(_ context.Context) (string, error) {
	return repo.remoteOriginUrl(repo.GetClonePath())
}

func (repo *Remote) IsEmpty(ctx context.Context) (bool, error) {
	return repo.isEmpty(ctx, repo.GetClonePath())
}

func (repo *Remote) IsShallowClone(ctx context.Context) (bool, error) {
	return true_git.IsShallowClone(ctx, repo.GetClonePath())
}

func (repo *Remote) IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	return true_git.IsAncestor(ctx, ancestorCommit, descendantCommit, repo.GetClonePath())
}

func (repo *Remote) CloneAndFetch(ctx context.Context) error {
	kind, err := repo.resolveMirrorKind()
	if err != nil {
		return err
	}
	repo.kind = kind

	logboek.Context(ctx).Debug().LogF("Using %s mirror for repo %q\n", kind, repo.String())

	if kind == mirrorKindShallow {
		return repo.cloneAndFetchShallow(ctx)
	}

	return repo.cloneAndFetchFull(ctx)
}

func (repo *Remote) cloneAndFetchFull(ctx context.Context) error {
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

func (repo *Remote) isCloneExistsForKind(kind mirrorKind) (bool, error) {
	_, err := os.Stat(repo.clonePathForKind(kind))
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

func buildCloneOptions(url, branch string) *git.CloneOptions {
	opts := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	if branch != "" {
		opts.SingleBranch = true
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		opts.Tags = git.NoTags
	}

	return opts
}

func buildFetchOptions(remoteName, branch string) *git.FetchOptions {
	opts := &git.FetchOptions{
		RemoteName: remoteName,
		Force:      true,
		Tags:       git.AllTags,
	}

	if branch != "" {
		opts.Tags = git.NoTags
		opts.RefSpecs = []gitconfig.RefSpec{gitconfig.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, remoteName, branch))}
	} else {
		// Explicit wildcard refspec: the mirror could have been cloned
		// single-branch by a branch mapping of the same URL, in which case its
		// configured refspec would never fetch commits outside that branch.
		opts.RefSpecs = []gitconfig.RefSpec{gitconfig.RefSpec(fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", remoteName))}
	}

	return opts
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

	kind := repo.mirrorKind()

	exists, err := repo.isCloneExistsForKind(kind)
	if err != nil {
		return false, err
	}
	if exists {
		if err := repo.updateLastAccessAt(ctx, repo.clonePathForKind(kind)); err != nil {
			return false, fmt.Errorf("error updating last access at timestamp: %w", err)
		}
		return false, nil
	}

	return true, repo.withMirrorKindLock(ctx, kind, func() error {
		return repo.cloneFullCore(ctx, kind)
	})
}

func (repo *Remote) cloneFullCore(ctx context.Context, kind mirrorKind) error {
	clonePath := repo.clonePathForKind(kind)

	exists, err := repo.isCloneExistsForKind(kind)
	if err != nil {
		return err
	}
	if exists {
		if err := repo.updateLastAccessAt(ctx, clonePath); err != nil {
			return fmt.Errorf("error updating last access at timestamp: %w", err)
		}

		return nil
	}

	logboek.Context(ctx).Default().LogFDetails("Clone %s\n", repo.Url)

	if err := os.MkdirAll(filepath.Dir(clonePath), 0o755); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", filepath.Dir(clonePath), err)
	}

	removeStaleCloneTmpDirs(ctx, clonePath)

	tmpPath := fmt.Sprintf("%s.%s.tmp", clonePath, uuid.New().String())
	defer os.RemoveAll(tmpPath)

	cloneOpts := buildCloneOptions(repo.Url, repo.Branch)

	if repo.BasicAuth != nil {
		cloneOpts.Auth = newBasicAuth(repo.BasicAuth.Username, repo.BasicAuth.Password).AuthMethod
	}

	if _, err := git.PlainCloneContext(ctx, tmpPath, true, cloneOpts); err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	if err := repo.updateLastAccessAt(ctx, tmpPath); err != nil {
		return fmt.Errorf("error updating last access at timestamp: %w", err)
	}

	peerWon, err := renameCloneIntoPlace(tmpPath, clonePath)
	if err != nil {
		return err
	}
	if peerWon {
		return repo.fetchOriginFullCore(ctx, kind)
	}

	return nil
}

const cloneTmpStalenessWindow = 3 * 24 * time.Hour

// removeStaleCloneTmpDirs reclaims tmp dirs abandoned by SIGKILLed clones.
// The age check makes the sweep safe even against a werf process that does
// not share our locker dir: no clone stays in progress for days.
func removeStaleCloneTmpDirs(ctx context.Context, clonePath string) {
	tmpPaths, err := filepath.Glob(clonePath + ".*")
	if err != nil {
		logboek.Context(ctx).Warn().LogF("Unable to look up stale tmp dirs of %q: %s\n", clonePath, err)
		return
	}

	for _, tmpPath := range tmpPaths {
		if !strings.HasSuffix(tmpPath, ".tmp") {
			continue
		}

		info, err := os.Stat(tmpPath)
		if err != nil || time.Since(info.ModTime()) <= cloneTmpStalenessWindow {
			continue
		}

		if err := os.RemoveAll(tmpPath); err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to remove stale tmp dir %q: %s\n", tmpPath, err)
		}
	}
}

// renameCloneIntoPlace treats a failed rename as a lost race when the
// destination already exists: a concurrent werf process (possibly another
// version sharing WERF_HOME) cloned the same mirror first. The peer's clone
// may have been made with a different refspec, so callers must fetch their
// own refs into it instead of treating it as their fresh clone.
func renameCloneIntoPlace(tmpPath, clonePath string) (bool, error) {
	renameErr := os.Rename(tmpPath, clonePath)
	if renameErr == nil {
		return false, nil
	}

	if _, err := os.Stat(clonePath); err == nil {
		return true, nil
	}

	return false, fmt.Errorf("rename %s to %s failed: %w", tmpPath, clonePath, renameErr)
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

	kind := repo.mirrorKind()

	return repo.withMirrorKindLock(ctx, kind, func() error {
		return repo.fetchOriginFullCore(ctx, kind)
	})
}

func (repo *Remote) fetchOriginFullCore(ctx context.Context, kind mirrorKind) error {
	clonePath := repo.clonePathForKind(kind)

	cfgPath := filepath.Join(clonePath, "config")

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

	rawRepo, err := gitRepoPlainOpen(clonePath)
	if err != nil {
		return fmt.Errorf("cannot open repo: %w", err)
	}

	logboek.Context(ctx).Default().LogFDetails("Fetch remote %s of %s\n", remoteName, repo.Url)

	fetchOpts := buildFetchOptions(remoteName, repo.Branch)

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

	commitHash, err := peelToCommitHash(rawRepo, ref.Hash())
	if err != nil {
		return "", fmt.Errorf("bad tag %q of repo %s: %w", tag, repo.String(), err)
	}

	res := commitHash.String()

	logboek.Context(ctx).Info().LogF("Using commit %q of repo %q tag %q\n", res, repo.String(), tag)

	return res, nil
}

func peelToCommitHash(rawRepo *git.Repository, hash plumbing.Hash) (plumbing.Hash, error) {
	obj, err := rawRepo.Object(plumbing.AnyObject, hash)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	for {
		switch typedObj := obj.(type) {
		case *object.Tag:
			if typedObj.TargetType != plumbing.CommitObject {
				obj, err = typedObj.Object()
				if err != nil {
					return plumbing.ZeroHash, err
				}

				continue
			}

			return typedObj.Target, nil
		case *object.Commit:
			return typedObj.Hash, nil
		default:
			return plumbing.ZeroHash, fmt.Errorf("unsupported tag target %q", typedObj.Type())
		}
	}
}

func (repo *Remote) GetOrCreatePatch(ctx context.Context, opts PatchOptions) (Patch, error) {
	return repo.getOrCreatePatch(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getRepoID(), repo.getWorkTreeCacheDir(repo.getRepoID()), opts)
}

func (repo *Remote) GetOrCreateChangedPaths(ctx context.Context, fromCommit, toCommit string) ([]true_git.ChangedPath, error) {
	return repo.getOrCreateChangedPaths(ctx, repo.GetClonePath(), fromCommit, toCommit)
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
	if repo.mirrorKind() == mirrorKindFull {
		return filepath.Join(GetWorkTreeCacheDir(), "remote", repoID)
	}
	return filepath.Join(GetWorkTreeCacheDir(), "remote", fmt.Sprintf("%s.%s", repoID, repo.mirrorKind()))
}

func (repo *Remote) withMirrorKindLock(ctx context.Context, kind mirrorKind, f func() error) error {
	opts := lockgate.AcquireOptions{Timeout: 600 * time.Second}
	repoIDLockName := fmt.Sprintf("remote_git.%s.%s", repo.getRepoID(), kind)

	if kind == mirrorKindFull {
		// remote_git_mapping.<name> is the lock name used by released werf 2.74.x
		// binaries for the shared full-mirror dir; matching it byte-for-byte is
		// what gives cross-version exclusion on hosts with a shared WERF_HOME.
		return werf.HostLocker().WithLock(ctx, fmt.Sprintf("remote_git_mapping.%s", repo.Name), opts, func() error {
			return werf.HostLocker().WithLock(ctx, repoIDLockName, opts, f)
		})
	}

	return werf.HostLocker().WithLock(ctx, repoIDLockName, opts, f)
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
