package stage_test

import (
	"context"
	"time"

	"github.com/werf/werf/v3/pkg/git_repo"
)

var _ git_repo.GitRepo = (*gitArchiveRepoStub)(nil)

type gitArchiveRepoStub struct {
	*GitRepoStub
	checksum string
	origin   string
	date     time.Time
}

func (repo *gitArchiveRepoStub) CommitMetadata(_ context.Context, commit string) (git_repo.CommitMetadata, error) {
	return git_repo.CommitMetadata{Hash: commit, CommitterDate: repo.date}, nil
}

func (repo *gitArchiveRepoStub) RemoteOriginUrl(_ context.Context) (string, error) {
	return repo.origin, nil
}

func (repo *gitArchiveRepoStub) TagCommit(_ context.Context, _ string) (string, error) {
	return repo.headCommitHash, nil
}

func (repo *gitArchiveRepoStub) LatestBranchCommit(_ context.Context, _ string) (string, error) {
	return repo.headCommitHash, nil
}

func (repo *gitArchiveRepoStub) IsCommitExists(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func (repo *gitArchiveRepoStub) GetOrCreateChecksum(_ context.Context, _ git_repo.ChecksumOptions) (string, error) {
	return repo.checksum, nil
}
