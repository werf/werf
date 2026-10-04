package git_repo

import (
	"context"

	"github.com/werf/werf/v3/pkg/git_repo/repo_handle"
)

// WithRepoHandleForTest exposes the repo handle scope to the external test
// package, which is the only place able to drive the production GitDataManager
// (pkg/git_repo/gitdata imports this package, so internal tests cannot).
func WithRepoHandleForTest(ctx context.Context, repo interface {
	withRepoHandle(ctx context.Context, commit string, f func(handle repo_handle.Handle) error) error
}, commit string, f func() error,
) error {
	return repo.withRepoHandle(ctx, commit, func(repo_handle.Handle) error {
		return f()
	})
}
