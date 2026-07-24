package git_repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
)

var fullLengthCommitSHARegexp = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

var errSubmodulesDetected = errors.New("submodules detected in commit")

type shallowUserError struct {
	err error
}

func (e shallowUserError) Error() string {
	return e.err.Error()
}

func (e shallowUserError) Unwrap() error {
	return e.err
}

var (
	lsRemoteTagsCacheMu sync.Mutex
	lsRemoteTagsCache   = map[string]map[string]true_git.RemoteTagRef{}
)

func resetLsRemoteTagsCache() {
	lsRemoteTagsCacheMu.Lock()
	defer lsRemoteTagsCacheMu.Unlock()
	lsRemoteTagsCache = map[string]map[string]true_git.RemoteTagRef{}
}

func (repo *Remote) cloneAndFetchShallow(ctx context.Context) error {
	if repo.IsDryRun {
		return nil
	}

	if repo.Commit != "" && !fullLengthCommitSHARegexp.MatchString(repo.Commit) {
		return fmt.Errorf("bad commit %q of repo %s: full-length commit SHA required", repo.Commit, repo.String())
	}

	if lock, err := CommonGitDataManager.LockGC(ctx, true); err != nil {
		return err
	} else {
		defer werf.HostLocker().ReleaseLock(lock)
	}

	return repo.withMirrorKindLock(ctx, mirrorKindShallow, func() error {
		shallowErr := repo.syncShallow(ctx)
		if shallowErr == nil {
			return nil
		}

		var userErr shallowUserError
		if errors.As(shallowErr, &userErr) {
			return shallowErr
		}

		persistMarker := errors.Is(shallowErr, errSubmodulesDetected) || isBySHAFetchRefusal(shallowErr)

		logboek.Context(ctx).Info().LogF("Falling back to full mirror for repo %q: %s\n", repo.String(), shallowErr)

		if err := repo.downgradeToFull(ctx, persistMarker); err != nil {
			return fmt.Errorf("shallow fetch failed and full mirror fallback also failed: %w; underlying shallow error: %v", err, shallowErr)
		}

		return nil
	})
}

// isBySHAFetchRefusal recognizes server-side policy refusal of by-SHA fetch
// (uploadpack.allowReachableSHA1InWant disabled) as opposed to transient errors.
func isBySHAFetchRefusal(err error) bool {
	msg := err.Error()
	for _, pattern := range []string{
		"unadvertised object",
		"not our ref",
		"does not support shallow",
	} {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

func (repo *Remote) syncShallow(ctx context.Context) error {
	markerExists, err := util.FileExists(repo.requiresFullMarkerPath())
	if err != nil {
		return fmt.Errorf("unable to check requires_full marker %q: %w", repo.requiresFullMarkerPath(), err)
	}
	if markerExists {
		return fmt.Errorf("requires_full marker is set for repo %s", repo.String())
	}

	if repo.Commit != "" {
		return repo.syncShallowCommit(ctx)
	}

	return repo.syncShallowTag(ctx)
}

func (repo *Remote) syncShallowCommit(ctx context.Context) error {
	shallowPath := repo.clonePathForKind(mirrorKindShallow)

	exists, err := repo.ensureShallowMirror(ctx)
	if err != nil {
		return err
	}

	if exists {
		commitExists, err := repo.isCommitExists(ctx, shallowPath, shallowPath, repo.Commit)
		if err != nil {
			return err
		}
		if commitExists {
			logboek.Context(ctx).Debug().LogF("Skipping fetch: commit %s already present in shallow mirror of repo %q\n", repo.Commit, repo.String())
			return repo.checkShallowCommitSubmodules(shallowPath, repo.Commit)
		}
	}

	logboek.Context(ctx).Default().LogFDetails("Fetch commit %s of %s (shallow)\n", repo.Commit, repo.Url)

	refSpec := fmt.Sprintf("+%s:refs/werf/commits/%s", repo.Commit, repo.Commit)
	if err := repo.shallowFetch(ctx, shallowPath, refSpec); err != nil {
		return err
	}

	commitExists, err := repo.isCommitExists(ctx, shallowPath, shallowPath, repo.Commit)
	if err != nil {
		return err
	}
	if !commitExists {
		return shallowUserError{fmt.Errorf("bad commit %q of repo %s: not found after shallow fetch", repo.Commit, repo.String())}
	}

	return repo.checkShallowCommitSubmodules(shallowPath, repo.Commit)
}

func (repo *Remote) syncShallowTag(ctx context.Context) error {
	shallowPath := repo.clonePathForKind(mirrorKindShallow)

	resolved, err := repo.lsRemoteTag(ctx, false)
	if err != nil {
		return err
	}

	if _, err := repo.ensureShallowMirror(ctx); err != nil {
		return err
	}

	localTagCommit, err := repo.localTagCommit(shallowPath)
	if err != nil {
		return err
	}

	if localTagCommit == resolved {
		logboek.Context(ctx).Debug().LogF("Skipping fetch: tag %q already resolves to commit %s in shallow mirror of repo %q\n", repo.Tag, resolved, repo.String())
		return repo.checkShallowCommitSubmodules(shallowPath, resolved)
	}

	commitExists, err := repo.isCommitExists(ctx, shallowPath, shallowPath, resolved)
	if err != nil {
		return err
	}
	if commitExists {
		logboek.Context(ctx).Debug().LogF("Updating tag ref %q to already present commit %s in shallow mirror of repo %q without fetch\n", repo.Tag, resolved, repo.String())
		if err := true_git.UpdateRef(ctx, shallowPath, "refs/tags/"+repo.Tag, resolved); err != nil {
			return err
		}
		return repo.checkShallowCommitSubmodules(shallowPath, resolved)
	}

	if err := repo.fetchShallowTagAndVerify(ctx, shallowPath, resolved); err != nil {
		return err
	}

	localTagCommit, err = repo.localTagCommit(shallowPath)
	if err != nil {
		return err
	}

	return repo.checkShallowCommitSubmodules(shallowPath, localTagCommit)
}

func (repo *Remote) fetchShallowTagAndVerify(ctx context.Context, shallowPath, expectedCommit string) error {
	logboek.Context(ctx).Default().LogFDetails("Fetch tag %s of %s (shallow)\n", repo.Tag, repo.Url)

	refSpec := fmt.Sprintf("+refs/tags/%s:refs/tags/%s", repo.Tag, repo.Tag)
	if err := repo.shallowFetch(ctx, shallowPath, refSpec); err != nil {
		return err
	}

	localTagCommit, err := repo.localTagCommit(shallowPath)
	if err != nil {
		return err
	}
	if localTagCommit == expectedCommit {
		return nil
	}

	refreshed, err := repo.lsRemoteTag(ctx, true)
	if err != nil {
		return err
	}
	if localTagCommit == refreshed {
		return nil
	}

	if err := repo.shallowFetch(ctx, shallowPath, refSpec); err != nil {
		return err
	}

	localTagCommit, err = repo.localTagCommit(shallowPath)
	if err != nil {
		return err
	}
	if localTagCommit != refreshed {
		return fmt.Errorf("tag %q of repo %s: fetched commit %s does not match remote commit %s", repo.Tag, repo.String(), localTagCommit, refreshed)
	}

	return nil
}

func (repo *Remote) localTagCommit(shallowPath string) (string, error) {
	rawRepo, err := gitRepoPlainOpen(shallowPath)
	if err != nil {
		return "", fmt.Errorf("cannot open repo %q: %w", shallowPath, err)
	}

	ref, err := rawRepo.Tag(repo.Tag)
	if errors.Is(err, git.ErrTagNotFound) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("bad tag %q of repo %s: %w", repo.Tag, repo.String(), err)
	}

	commitHash, err := peelToCommitHash(rawRepo, ref.Hash())
	if err != nil {
		return "", fmt.Errorf("bad tag %q of repo %s: %w", repo.Tag, repo.String(), err)
	}

	return commitHash.String(), nil
}

// lsRemoteTag resolves the peeled commit SHA of repo.Tag via a per-URL cached
// `git ls-remote --tags` call, batched to at most one network call per werf
// process per URL (unless fresh is requested).
func (repo *Remote) lsRemoteTag(ctx context.Context, fresh bool) (string, error) {
	lsRemoteTagsCacheMu.Lock()
	defer lsRemoteTagsCacheMu.Unlock()

	cacheKey := repo.lsRemoteTagsCacheKey()

	tags, cached := lsRemoteTagsCache[cacheKey]
	if !cached || fresh {
		env, cleanup, err := basicAuthEnv(repo.BasicAuth)
		if err != nil {
			return "", err
		}
		defer cleanup()

		tags, err = true_git.LsRemoteTags(ctx, repo.Url, true_git.LsRemoteTagsOptions{Env: env})
		if err != nil {
			return "", fmt.Errorf("cannot list remote tags of repo %q: %w", repo.String(), err)
		}

		lsRemoteTagsCache[cacheKey] = tags
	}

	tagRef, found := tags[repo.Tag]
	if !found {
		return "", shallowUserError{fmt.Errorf("bad tag %q of repo %s: tag not found in remote", repo.Tag, repo.String())}
	}

	if tagRef.PeeledSHA != "" {
		return tagRef.PeeledSHA, nil
	}
	return tagRef.ObjectSHA, nil
}

func (repo *Remote) lsRemoteTagsCacheKey() string {
	var user, password string
	if repo.BasicAuth != nil {
		user = repo.BasicAuth.Username
		password = repo.BasicAuth.Password
	}
	return util.Sha256Hash(repo.Url, user, password)
}

func (repo *Remote) shallowFetch(ctx context.Context, shallowPath, refSpec string) error {
	env, cleanup, err := basicAuthEnv(repo.BasicAuth)
	if err != nil {
		return err
	}
	defer cleanup()

	return true_git.ShallowFetch(ctx, shallowPath, []string{refSpec}, true_git.ShallowFetchOptions{Env: env})
}

func (repo *Remote) checkShallowCommitSubmodules(shallowPath, commit string) error {
	rawRepo, err := gitRepoPlainOpen(shallowPath)
	if err != nil {
		return fmt.Errorf("cannot open repo %q: %w", shallowPath, err)
	}

	commitHash, err := newHash(commit)
	if err != nil {
		return fmt.Errorf("bad commit hash %q: %w", commit, err)
	}

	commitObj, err := rawRepo.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("bad commit %q: %w", commit, err)
	}

	hasSubmodules, err := HasSubmodulesInCommit(commitObj)
	if err != nil {
		return err
	}
	if hasSubmodules {
		return errSubmodulesDetected
	}

	return nil
}

func (repo *Remote) ensureShallowMirror(ctx context.Context) (bool, error) {
	shallowPath := repo.clonePathForKind(mirrorKindShallow)

	exists, err := repo.isCloneExistsForKind(mirrorKindShallow)
	if err != nil {
		return false, err
	}
	if exists {
		if err := repo.updateLastAccessAt(ctx, shallowPath); err != nil {
			return false, fmt.Errorf("error updating last access at timestamp: %w", err)
		}
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(shallowPath), 0o755); err != nil {
		return false, fmt.Errorf("unable to create dir %s: %w", filepath.Dir(shallowPath), err)
	}

	tmpPath := fmt.Sprintf("%s.tmp", shallowPath)
	if err := os.RemoveAll(tmpPath); err != nil {
		return false, fmt.Errorf("unable to prepare tmp path %s: failed to remove: %w", tmpPath, err)
	}
	defer os.RemoveAll(tmpPath)

	if err := true_git.InitBareRepoWithOrigin(ctx, tmpPath, repo.Url); err != nil {
		return false, err
	}

	if err := repo.updateLastAccessAt(ctx, tmpPath); err != nil {
		return false, fmt.Errorf("error updating last access at timestamp: %w", err)
	}

	if err := os.Rename(tmpPath, shallowPath); err != nil {
		return false, fmt.Errorf("rename %s to %s failed: %w", tmpPath, shallowPath, err)
	}

	return false, nil
}

// downgradeToFull is entered while the shallow mirror kind lock is held. It
// prepares the full mirror under the full kind lock, optionally persists the
// requires_full marker (only for capability/submodule-driven downgrades, and
// only after the full mirror is confirmed usable), and switches repo.kind.
func (repo *Remote) downgradeToFull(ctx context.Context, persistMarker bool) error {
	return repo.withMirrorKindLock(ctx, mirrorKindFull, func() error {
		exists, err := repo.isCloneExistsForKind(mirrorKindFull)
		if err != nil {
			return err
		}

		if exists {
			if err := repo.updateLastAccessAt(ctx, repo.clonePathForKind(mirrorKindFull)); err != nil {
				return fmt.Errorf("error updating last access at timestamp: %w", err)
			}
			if err := repo.fetchOriginFullCore(ctx, mirrorKindFull); err != nil {
				return err
			}
		} else {
			if err := repo.cloneFullCore(ctx, mirrorKindFull); err != nil {
				return err
			}

			rawRepo, err := gitRepoPlainOpen(repo.clonePathForKind(mirrorKindFull))
			if err != nil {
				return fmt.Errorf("open cloned repo: %w", err)
			}
			if err := repo.syncLocalBranches(ctx, rawRepo); err != nil {
				return err
			}
		}

		if err := repo.verifyTargetInFullMirror(ctx, repo.clonePathForKind(mirrorKindFull)); err != nil {
			return err
		}

		if persistMarker {
			if err := repo.writeRequiresFullMarker(); err != nil {
				return err
			}

			shallowPath := repo.clonePathForKind(mirrorKindShallow)
			if err := os.RemoveAll(shallowPath); err != nil {
				return fmt.Errorf("unable to remove stale shallow mirror %q: %w", shallowPath, err)
			}
			if err := os.RemoveAll(repo.getWorkTreeCacheDir(repo.getRepoID())); err != nil {
				return fmt.Errorf("unable to remove stale shallow worktree cache: %w", err)
			}
		}

		repo.kind = mirrorKindFull

		return nil
	})
}

func (repo *Remote) verifyTargetInFullMirror(ctx context.Context, fullPath string) error {
	if repo.Commit != "" {
		exists, err := repo.isCommitExists(ctx, fullPath, fullPath, repo.Commit)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("bad commit %q of repo %s: not found in full mirror after fallback", repo.Commit, repo.String())
		}
		return nil
	}

	if repo.Tag != "" {
		tagCommit, err := repo.localTagCommit(fullPath)
		if err != nil {
			return err
		}
		if tagCommit == "" {
			return fmt.Errorf("bad tag %q of repo %s: not found in full mirror after fallback", repo.Tag, repo.String())
		}
	}

	return nil
}

func (repo *Remote) writeRequiresFullMarker() error {
	markerPath := repo.requiresFullMarkerPath()

	tmpPath := fmt.Sprintf("%s.tmp", markerPath)
	if err := os.WriteFile(tmpPath, nil, 0o644); err != nil {
		return fmt.Errorf("unable to write requires_full marker %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, markerPath); err != nil {
		return fmt.Errorf("rename %s to %s failed: %w", tmpPath, markerPath, err)
	}

	return nil
}

// basicAuthEnv builds shell git env for basic auth via a temp GIT_ASKPASS
// helper so that credentials never appear in process argv or on-disk config.
func basicAuthEnv(auth *BasicAuth) ([]string, func(), error) {
	env := []string{"GIT_TERMINAL_PROMPT=0"}

	if auth == nil {
		return env, func() {}, nil
	}

	dir, err := os.MkdirTemp("", "werf-git-askpass-")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create askpass temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(dir) }

	script := filepath.Join(dir, "askpass.sh")
	content := `#!/bin/sh
case "$1" in
Username*) printf '%s' "$GIT_ASKPASS_USERNAME" ;;
*) printf '%s' "$GIT_ASKPASS_PASSWORD" ;;
esac
`
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("unable to write askpass script %q: %w", script, err)
	}

	env = append(env,
		"GIT_ASKPASS="+script,
		"GIT_ASKPASS_USERNAME="+auth.Username,
		"GIT_ASKPASS_PASSWORD="+auth.Password,
	)

	return env, cleanup, nil
}
