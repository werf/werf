package git_repo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	"github.com/werf/werf/v3/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v3/pkg/opstats"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
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

	resolvedRefsMutex sync.Mutex
	resolvedRefs      map[string]string
}

// Resolving the mapped branch or tag twice in one build must give the same
// answer even if the origin moved meanwhile: the SHA resolved first is what
// every later stage, digest and cache lookup of this build is built on. Only an
// explicit fetch may change it.
func (repo *Remote) resolvedRef(key string) (string, bool) {
	repo.resolvedRefsMutex.Lock()
	defer repo.resolvedRefsMutex.Unlock()

	commit, ok := repo.resolvedRefs[key]
	return commit, ok
}

func (repo *Remote) rememberResolvedRef(key, commit string) {
	repo.resolvedRefsMutex.Lock()
	defer repo.resolvedRefsMutex.Unlock()

	if repo.resolvedRefs == nil {
		repo.resolvedRefs = map[string]string{}
	}
	repo.resolvedRefs[key] = commit
}

func (repo *Remote) forgetResolvedRefs() {
	repo.resolvedRefsMutex.Lock()
	defer repo.resolvedRefsMutex.Unlock()

	repo.resolvedRefs = nil
}

// mappedCommit is the SHA this repo is pinned to: the mapped commit, or the one
// already resolved for the mapped tag or branch. Restoring an evicted mirror
// must bring that SHA back instead of the current tip.
func (repo *Remote) mappedCommit() string {
	if repo.Commit != "" {
		return repo.Commit
	}

	if repo.Tag != "" {
		if commit, ok := repo.resolvedRef(resolvedTagRefKey(repo.Tag)); ok {
			return commit
		}
	}

	if repo.Branch != "" {
		if commit, ok := repo.resolvedRef(resolvedBranchRefKey(repo.Branch)); ok {
			return commit
		}
	}

	return ""
}

func resolvedBranchRefKey(branch string) string {
	return "branch:" + branch
}

func resolvedTagRefKey(tag string) string {
	return "tag:" + tag
}

func OpenRemoteRepo(name, url string, auth *BasicAuthCredentials) (*Remote, error) {
	repo := &Remote{Url: url}
	repo.Base = NewBase(name, repo.initRepoHandleBackedByWorkTree)
	repo.Base.ensureRepoDataFunc = repo.ensureMirrorData
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
	var res string
	err := repo.withMirror(ctx, toCommit, func() (err error) {
		res, err = repo.createDetachedMergeCommit(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getWorkTreeCacheDir(repo.getRepoID()), fromCommit, toCommit)
		return err
	})
	return res, err
}

func (repo *Remote) GetMergeCommitParents(ctx context.Context, commit string) ([]string, error) {
	var res []string
	err := repo.withMirror(ctx, commit, func() (err error) {
		res, err = repo.getMergeCommitParents(repo.GetClonePath(), commit)
		return err
	})
	return res, err
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

func (repo *Remote) RemoteOriginUrl(ctx context.Context) (string, error) {
	var res string
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.remoteOriginUrl(repo.GetClonePath())
		return err
	})
	return res, err
}

func (repo *Remote) IsEmpty(ctx context.Context) (bool, error) {
	var res bool
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.isEmpty(ctx, repo.GetClonePath())
		return err
	})
	return res, err
}

func (repo *Remote) IsShallowClone(ctx context.Context) (bool, error) {
	var res bool
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = true_git.IsShallowClone(ctx, repo.GetClonePath())
		return err
	})
	return res, err
}

func (repo *Remote) IsAncestor(ctx context.Context, ancestorCommit, descendantCommit string) (bool, error) {
	var res bool
	err := repo.withMirror(ctx, descendantCommit, func() (err error) {
		res, err = true_git.IsAncestor(ctx, ancestorCommit, descendantCommit, repo.GetClonePath())
		return err
	})
	return res, err
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

	doneClone := opstats.Observe(ctx, opstats.OperationGitClone)
	defer doneClone()

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

	if err := timestamps.WriteTimestampFile(filepath.Join(tmpPath, "last_access_at"), time.Now()); err != nil {
		return fmt.Errorf("error writing last access at timestamp: %w", err)
	}

	peerWon, err := renameCloneIntoPlace(tmpPath, clonePath)
	if err != nil {
		return err
	}
	if peerWon {
		// Stop the clone observation before the fetch so git fetch is not
		// recorded inside the git clone interval.
		doneClone()
		return repo.fetchOriginFullCore(ctx, kind)
	}

	return nil
}

// cloneTmpStalenessWindow encodes the same judgement as gitdata's
// cacheVersionStalenessWindow: how long a werf process can plausibly still be
// mid-operation. Tune them together.
const cloneTmpStalenessWindow = 3 * 24 * time.Hour

// removeStaleCloneTmpDirs reclaims tmp dirs abandoned by SIGKILLed clones. A
// tmp dir is swept only when nothing in its subtree was written within the
// window (unlike gitdata's hasFreshFiles, directory mtimes count — keeping is
// the safe direction here), which makes the sweep safe even against a werf
// process that does not share our locker dir: a live clone writes
// continuously.
func removeStaleCloneTmpDirs(ctx context.Context, clonePath string) {
	parentDir := filepath.Dir(clonePath)

	entries, err := os.ReadDir(parentDir)
	if err != nil {
		logboek.Context(ctx).Warn().LogF("Unable to look up stale tmp dirs of %q: %s\n", clonePath, err)
		return
	}

	prefix := filepath.Base(clonePath) + "."

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix) || !strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}

		tmpPath := filepath.Join(parentDir, entry.Name())
		if !isCloneTmpDirAbandoned(tmpPath) {
			continue
		}

		if err := os.RemoveAll(tmpPath); err != nil {
			logboek.Context(ctx).Warn().LogF("Unable to remove stale tmp dir %q: %s\n", tmpPath, err)
		}
	}
}

func isCloneTmpDirAbandoned(tmpPath string) bool {
	abandoned := true

	err := filepath.WalkDir(tmpPath, func(_ string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := dirEntry.Info()
		if err != nil {
			return err
		}

		if time.Since(info.ModTime()) <= cloneTmpStalenessWindow {
			abandoned = false
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return false
	}

	return abandoned
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

	// An explicit fetch is the one place allowed to change what the mapped
	// branch or tag resolves to, so the memo is dropped before the mirror is
	// ensured: the fetch must not be pinned to the SHA it is about to replace.
	repo.forgetResolvedRefs()

	kind := repo.mirrorKind()

	// Clone released the GC lock before returning, so the mirror this fetch
	// targets can already be gone by now.
	return repo.withMirror(ctx, "", func() error {
		return repo.withMirrorKindLock(ctx, kind, func() error {
			return repo.fetchOriginFullCore(ctx, kind)
		})
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

	defer opstats.Observe(ctx, opstats.OperationGitFetch)()

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
	var res string
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = getHeadCommit(ctx, repo.GetClonePath())
		return err
	})
	return res, err
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
	if commit, ok := repo.resolvedRef(resolvedBranchRefKey(branch)); ok {
		return commit, nil
	}

	var res string
	if err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.latestBranchCommit(ctx, branch)
		return err
	}); err != nil {
		return "", err
	}

	repo.rememberResolvedRef(resolvedBranchRefKey(branch), res)

	return res, nil
}

func (repo *Remote) latestBranchCommit(ctx context.Context, branch string) (string, error) {
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
	if commit, ok := repo.resolvedRef(resolvedTagRefKey(tag)); ok {
		return commit, nil
	}

	var res string
	if err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.tagCommit(ctx, tag)
		return err
	}); err != nil {
		return "", err
	}

	repo.rememberResolvedRef(resolvedTagRefKey(tag), res)

	return res, nil
}

func (repo *Remote) tagCommit(ctx context.Context, tag string) (string, error) {
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
	var res Patch
	err := repo.withMirror(ctx, opts.ToCommit, func() (err error) {
		res, err = repo.getOrCreatePatch(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getRepoID(), repo.getWorkTreeCacheDir(repo.getRepoID()), opts)
		return err
	})
	return res, err
}

func (repo *Remote) GetOrCreateChangedPaths(ctx context.Context, fromCommit, toCommit string) ([]true_git.ChangedPath, error) {
	var res []true_git.ChangedPath
	err := repo.withMirror(ctx, toCommit, func() (err error) {
		res, err = repo.getOrCreateChangedPaths(ctx, repo.GetClonePath(), fromCommit, toCommit)
		return err
	})
	return res, err
}

func (repo *Remote) GetOrCreateArchive(ctx context.Context, opts ArchiveOptions) (Archive, error) {
	var res Archive
	err := repo.withMirror(ctx, opts.Commit, func() (err error) {
		res, err = repo.getOrCreateArchive(ctx, repo.GetClonePath(), repo.GetClonePath(), repo.getRepoID(), repo.getWorkTreeCacheDir(repo.getRepoID()), opts)
		return err
	})
	return res, err
}

func (repo *Remote) GetOrCreateChecksum(ctx context.Context, opts ChecksumOptions) (checksum string, err error) {
	err = repo.withRepoHandle(ctx, opts.Commit, func(repoHandle repo_handle.Handle) error {
		checksum, err = repo.getOrCreateChecksum(ctx, repoHandle, opts)
		return err
	})

	return checksum, err
}

// IsCommitExists answers from the restored mirror: a missing commit is a
// legitimate answer here, so it is not passed as a pin to withMirror.
func (repo *Remote) IsCommitExists(ctx context.Context, commit string) (bool, error) {
	var res bool
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.isCommitExists(ctx, repo.GetClonePath(), repo.GetClonePath(), commit)
		return err
	})
	return res, err
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

func (repo *Remote) TagsList(ctx context.Context) ([]string, error) {
	var res []string
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.tagsList(repo.GetClonePath())
		return err
	})
	return res, err
}

func (repo *Remote) RemoteBranchesList(ctx context.Context) ([]string, error) {
	var res []string
	err := repo.withMirror(ctx, "", func() (err error) {
		res, err = repo.remoteBranchesList(repo.GetClonePath())
		return err
	})
	return res, err
}

// withMirror runs f with the mirror present and protected from GC. commit is
// the SHA a caller already resolved, or "" when f resolves one itself.
func (repo *Remote) withMirror(ctx context.Context, commit string, f func() error) error {
	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return err
	} else {
		defer werf.HostLocker().ReleaseLock(lock)
	}

	if _, err := repo.ensureMirrorData(ctx, commit); err != nil {
		return err
	}

	return f()
}

// ensureMirrorData refreshes last_access_at (reads, not only clones, keep a
// mirror in use) and restores a mirror an already running GC removed. The
// caller must hold the shared GC lock. It reports whether a restore happened.
func (repo *Remote) ensureMirrorData(ctx context.Context, commit string) (bool, error) {
	if repo.IsDryRun {
		return false, nil
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

		pin := commit
		if pin == "" {
			pin = repo.mappedCommit()
		}
		if pin == "" {
			return false, nil
		}

		// The mirror dir being there does not mean it is still ours: GC may have
		// taken it and a peer may have cloned a force-pushed origin into the same
		// path, without the commit this build is pinned to.
		pinExists, err := repo.isCommitExists(ctx, repo.clonePathForKind(kind), repo.clonePathForKind(kind), pin)
		if err != nil {
			return false, err
		}
		if pinExists {
			return false, nil
		}

		if err := repo.withMirrorKindLock(ctx, kind, func() error {
			return repo.recoverPinnedCommit(ctx, kind, pin)
		}); err != nil {
			return false, fmt.Errorf("recover commit %s in %s mirror of repo %q: %w", pin, kind, repo.String(), err)
		}

		return true, nil
	}

	logboek.Context(ctx).Warn().LogF("WARNING: The %s mirror of repo %q is gone from the local cache, restoring it\n", kind, repo.String())

	if err := repo.withMirrorKindLock(ctx, kind, func() error {
		return repo.restoreMirror(ctx, kind, commit)
	}); err != nil {
		return false, fmt.Errorf("restore %s mirror of repo %q: %w", kind, repo.String(), err)
	}

	return true, nil
}

// restoreMirror re-creates the mirror around commit, which is immutable input:
// a branch or tag that moved meanwhile must not be substituted for it. With an
// empty commit the mapping is restored as the initial clone and fetch would.
func (repo *Remote) restoreMirror(ctx context.Context, kind mirrorKind, commit string) error {
	if commit == "" {
		commit = repo.mappedCommit()
	}

	if kind == mirrorKindShallow {
		if commit == "" {
			return repo.syncShallow(ctx)
		}
		return repo.restoreShallowMirror(ctx, commit)
	}

	if err := repo.cloneFullCore(ctx, kind); err != nil {
		return err
	}

	clonePath := repo.clonePathForKind(kind)

	rawRepo, err := gitRepoPlainOpen(clonePath)
	if err != nil {
		return fmt.Errorf("open restored repo: %w", err)
	}
	if err := repo.syncLocalBranches(ctx, rawRepo); err != nil {
		return err
	}

	if commit == "" {
		return nil
	}

	commitExists, err := repo.isCommitExists(ctx, clonePath, clonePath, commit)
	if err != nil {
		return err
	}
	if commitExists {
		return nil
	}

	if err := repo.fetchOriginFullCore(ctx, kind); err != nil {
		return err
	}

	commitExists, err = repo.isCommitExists(ctx, clonePath, clonePath, commit)
	if err != nil {
		return err
	}
	if commitExists {
		return nil
	}

	// A force-pushed commit is reachable from no advertised ref, so ask the
	// origin for the SHA itself before giving up. Servers may refuse it.
	if err := repo.fetchCommitIntoFullMirror(ctx, clonePath, commit); err != nil {
		return fmt.Errorf("commit %s is not available in origin %s: %w", commit, repo.Url, err)
	}

	commitExists, err = repo.isCommitExists(ctx, clonePath, clonePath, commit)
	if err != nil {
		return err
	}
	if !commitExists {
		return fmt.Errorf("commit %s is not available in origin %s anymore", commit, repo.Url)
	}

	return nil
}

// recoverPinnedCommit fetches the pinned commit into a mirror that lost it. The
// commit is asked for by SHA: it is reachable from no ref of the force-pushed
// origin, and re-resolving the mapped ref would silently change the build.
func (repo *Remote) recoverPinnedCommit(ctx context.Context, kind mirrorKind, commit string) error {
	clonePath := repo.clonePathForKind(kind)

	pinExists, err := repo.isCommitExists(ctx, clonePath, clonePath, commit)
	if err != nil {
		return err
	}
	if pinExists {
		return nil
	}

	logboek.Context(ctx).Warn().LogF("WARNING: Commit %s is missing from the %s mirror of repo %q, fetching it again\n", commit, kind, repo.String())

	if kind == mirrorKindShallow {
		err = repo.shallowFetch(ctx, clonePath, fmt.Sprintf("+%s:refs/werf/commits/%s", commit, commit))
	} else {
		err = repo.fetchCommitIntoFullMirror(ctx, clonePath, commit)
	}
	if err != nil {
		return fmt.Errorf("commit %s is not available in origin %s: %w", commit, repo.Url, err)
	}

	pinExists, err = repo.isCommitExists(ctx, clonePath, clonePath, commit)
	if err != nil {
		return err
	}
	if !pinExists {
		return fmt.Errorf("commit %s is not available in origin %s anymore", commit, repo.Url)
	}

	return nil
}

func (repo *Remote) fetchCommitIntoFullMirror(ctx context.Context, clonePath, commit string) error {
	defer opstats.Observe(ctx, opstats.OperationGitFetch)()

	env, cleanup, err := basicAuthEnv(repo.BasicAuth)
	if err != nil {
		return err
	}
	defer cleanup()

	return true_git.Fetch(ctx, clonePath, true_git.FetchOptions{
		Env:      env,
		RefSpecs: map[string][]string{"origin": {fmt.Sprintf("+%s:refs/werf/commits/%s", commit, commit)}},
	})
}

func (repo *Remote) restoreShallowMirror(ctx context.Context, commit string) error {
	shallowPath := repo.clonePathForKind(mirrorKindShallow)

	if _, err := repo.ensureShallowMirror(ctx); err != nil {
		return err
	}

	commitExists, err := repo.isCommitExists(ctx, shallowPath, shallowPath, commit)
	if err != nil {
		return err
	}
	if commitExists {
		return nil
	}

	// Fetching the commit itself, not the mapped tag, is what pins the result:
	// refs/werf/commits/<commit> also anchors the objects against git's own gc.
	if err := repo.shallowFetch(ctx, shallowPath, fmt.Sprintf("+%s:refs/werf/commits/%s", commit, commit)); err != nil {
		return err
	}

	commitExists, err = repo.isCommitExists(ctx, shallowPath, shallowPath, commit)
	if err != nil {
		return err
	}
	if !commitExists {
		return fmt.Errorf("commit %s is not available in origin %s anymore", commit, repo.Url)
	}

	return nil
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
	if !hasSubmodules {
		return repo_handle.NewHandleWithoutSubmodules(repository), nil
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
