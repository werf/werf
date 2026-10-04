package includes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/git_repo"
	"github.com/werf/werf/v3/pkg/git_repo/gitdata"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

// setupRemoteInclude boots the production git cache in a temp WERF_HOME, makes a
// source repo with one commit and returns the include config plus the
// initialized remote repo cache. The origin is a file:// URL so the include is
// treated as remote (a plain path would be opened as a local repo).
func setupRemoteInclude(ctx context.Context, t *testing.T) (Config, *gitRepositoriesWithCache, string) {
	t.Helper()

	// test/pkg/utils command helpers assert via gomega
	gomega.RegisterTestingT(t)

	home := t.TempDir()
	if err := werf.Init(home, t.TempDir()); err != nil {
		t.Fatalf("werf init: %v", err)
	}
	if err := true_git.Init(ctx, true_git.Options{}); err != nil {
		t.Fatalf("true_git init: %v", err)
	}
	manager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		t.Fatalf("git data manager: %v", err)
	}
	if err := git_repo.Init(manager); err != nil {
		t.Fatalf("git_repo init: %v", err)
	}

	sourceDir := filepath.Join(home, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	utils.RunSucceedCommand(ctx, sourceDir, "git", "-c", "init.defaultBranch=main", "init")
	utils.RunSucceedCommand(ctx, sourceDir, "git", "config", "uploadpack.allowAnySHA1InWant", "true")
	utils.WriteFile(filepath.Join(sourceDir, "data.txt"), []byte("v1"))
	utils.RunSucceedCommand(ctx, sourceDir, "git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "add", "--", "data.txt")
	utils.RunSucceedCommand(ctx, sourceDir, "git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "v1")
	commit := utils.GetHeadCommit(ctx, sourceDir)

	cfg := Config{Includes: []includeConf{{
		Git:    "file://" + sourceDir,
		Branch: "main",
		Add:    "/",
		To:     "/",
	}}}

	repos, err := initRemoteRepos(ctx, cfg)
	if err != nil {
		t.Fatalf("init remote repos: %v", err)
	}

	return cfg, repos, commit
}

func remoteOf(t *testing.T, repos *gitRepositoriesWithCache, git string) *git_repo.Remote {
	t.Helper()
	r, err := repos.getRepository(git)
	if err != nil {
		t.Fatalf("get repository: %v", err)
	}
	remote, ok := r.repo.(*git_repo.Remote)
	if !ok {
		t.Fatalf("expected a remote repository, got %T", r.repo)
	}
	return remote
}

// GetIncludes walks the commit tree lazily, so an eviction between lock
// resolution and the walk used to break it: the mirror must be restored and
// held for the whole walk.
func TestGetIncludesSurvivesMirrorEviction(t *testing.T) {
	ctx := context.Background()
	cfg, repos, commit := setupRemoteInclude(ctx, t)

	lockInfo, err := getLockInfo(ctx, getLockInfoOptions{
		includesConfig:   cfg,
		useLatestVersion: true,
		remoteRepos:      repos,
	})
	if err != nil {
		t.Fatalf("get lock info: %v", err)
	}
	if got, err := lockInfo.GetCommit(cfg.Includes[0].Git, "main"); err != nil || got != commit {
		t.Fatalf("locked commit = %q, %v; want %q", got, err, commit)
	}

	// GC evicts the mirror after the lock file was built.
	if err := os.RemoveAll(remoteOf(t, repos, cfg.Includes[0].Git).GetClonePath()); err != nil {
		t.Fatalf("evict mirror: %v", err)
	}

	includes, err := GetIncludes(ctx, cfg, lockInfo, repos)
	if err != nil {
		t.Fatalf("get includes: %v", err)
	}
	if len(includes) != 1 {
		t.Fatalf("got %d includes, want 1", len(includes))
	}
	if includes[0].commitHash != commit {
		t.Fatalf("include commit = %q, want %q", includes[0].commitHash, commit)
	}

	data, err := includes[0].GetFile(ctx, "data.txt")
	if err != nil {
		t.Fatalf("read include file: %v", err)
	}
	if string(data) != "v1" {
		t.Fatalf("include file = %q, want %q", data, "v1")
	}
}

// The handle must be covered by the shared GC lock, not merely restored before
// use: GC must not be able to start while a callback still reads the mirror.
func TestWithRepositoryBlocksGCUntilCallbackReturns(t *testing.T) {
	ctx := context.Background()
	cfg, repos, _ := setupRemoteInclude(ctx, t)

	r, err := repos.getRepository(cfg.Includes[0].Git)
	if err != nil {
		t.Fatalf("get repository: %v", err)
	}

	gcEntered := make(chan struct{})
	inside := make(chan struct{})
	release := make(chan struct{})

	go func() {
		<-inside
		lock, err := git_repo.CommonGitDataManager.LockGC(ctx, false)
		if err != nil {
			return
		}
		close(gcEntered)
		_ = werf.HostLocker().ReleaseLock(lock)
	}()

	if err := r.withRepository(ctx, "", func(repo *git.Repository) error {
		close(inside)
		select {
		case <-gcEntered:
			t.Error("GC acquired the exclusive lock while the repository handle was in use")
		case <-time.After(500 * time.Millisecond):
		}
		close(release)
		return nil
	}); err != nil {
		t.Fatalf("with repository: %v", err)
	}

	select {
	case <-gcEntered:
	case <-time.After(30 * time.Second):
		t.Fatal("GC did not get the exclusive lock after the callback returned")
	}
}
